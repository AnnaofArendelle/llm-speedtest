// Package store persists endpoints, runs and measurements in SQLite
// (pure-Go driver). Writes are serialised and batched; reads use the
// connection pool with WAL mode.
package store

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite" // sqlite driver

	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/family"
	"github.com/AnnaofArendelle/llm-speedtest/internal/provider"
)

//go:embed migrations/*.sql
var migrations embed.FS

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("not found")

// KeyCodec encrypts API keys at rest.
type KeyCodec interface {
	Encrypt(plain string) ([]byte, error)
	Decrypt(blob []byte) (string, error)
}

// Store is the SQLite-backed persistence layer.
type Store struct {
	db      *sql.DB
	path    string
	writeMu sync.Mutex
	codec   KeyCodec
}

// Open opens (and migrates) the database at path.
func Open(path string, codec KeyCodec) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	db.SetConnMaxLifetime(0)
	s := &Store{db: db, path: path, codec: codec}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// Path returns the database file path.
func (s *Store) Path() string { return s.path }

func (s *Store) migrate() error {
	if _, err := s.db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return err
	}
	applied := map[int]bool{}
	rows, err := s.db.QueryContext(context.Background(), `SELECT version FROM schema_version`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			_ = rows.Close()
			return err
		}
		applied[v] = true
	}
	_ = rows.Close()
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		var v int
		if _, err := fmt.Sscanf(name, "%d_", &v); err != nil {
			return fmt.Errorf("bad migration name %s", name)
		}
		if applied[v] {
			continue
		}
		sqlText, err := migrations.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		tx, err := s.db.BeginTx(context.Background(), nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(context.Background(), string(sqlText)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(context.Background(), `INSERT INTO schema_version(version, applied_at) VALUES (?, ?)`, v, now()); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func parseTime(s sql.NullString) time.Time {
	if !s.Valid || s.String == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339Nano, s.String)
	return t
}

func mustJSON(v any) string {
	if v == nil {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// NewID returns a sortable unique id: timestamp + random suffix.
func NewID() string {
	var b [6]byte
	_, _ = randRead(b[:])
	return fmt.Sprintf("%s-%x", time.Now().UTC().Format("20060102-150405"), b)
}

// ---------------------------------------------------------------- endpoints

// Endpoint is a saved endpoint profile. APIKey is populated only by
// GetEndpointWithKey; listings carry KeyHint instead.
type Endpoint struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	BaseURL   string            `json:"base_url"`
	Dialect   string            `json:"dialect"`
	Headers   map[string]string `json:"headers"`
	APIKey    string            `json:"-"`
	KeyHint   string            `json:"key_hint"`
	HasKey    bool              `json:"has_key"`
	Insecure  bool              `json:"insecure"`
	Proxy     string            `json:"proxy"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// SaveEndpoint inserts or updates an endpoint. When ep.APIKey is empty the
// stored key is kept.
func (s *Store) SaveEndpoint(ep *Endpoint) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if ep.ID == "" {
		ep.ID = NewID()
		ep.CreatedAt = time.Now().UTC()
	}
	ep.UpdatedAt = time.Now().UTC()
	if ep.Headers == nil {
		ep.Headers = map[string]string{}
	}
	var keyEnc []byte
	keyHint := ep.KeyHint
	if ep.APIKey != "" {
		if s.codec == nil {
			return errors.New("cannot store API key: no key codec configured")
		}
		enc, err := s.codec.Encrypt(ep.APIKey)
		if err != nil {
			return err
		}
		keyEnc = enc
		keyHint = maskKey(ep.APIKey)
	}
	created := ep.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	if keyEnc != nil {
		_, err := s.db.ExecContext(context.Background(), `INSERT INTO endpoints(id,name,base_url,dialect,headers_json,key_enc,key_hint,insecure,proxy,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET name=excluded.name, base_url=excluded.base_url, dialect=excluded.dialect, headers_json=excluded.headers_json,
			key_enc=excluded.key_enc, key_hint=excluded.key_hint, insecure=excluded.insecure, proxy=excluded.proxy, updated_at=excluded.updated_at`,
			ep.ID, ep.Name, ep.BaseURL, ep.Dialect, mustJSON(ep.Headers), keyEnc, keyHint, boolInt(ep.Insecure), ep.Proxy, created.Format(time.RFC3339Nano), ep.UpdatedAt.Format(time.RFC3339Nano))
		if err != nil {
			return err
		}
	} else {
		_, err := s.db.ExecContext(context.Background(), `INSERT INTO endpoints(id,name,base_url,dialect,headers_json,key_enc,key_hint,insecure,proxy,created_at,updated_at)
			VALUES(?,?,?,?,?,NULL,'',?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET name=excluded.name, base_url=excluded.base_url, dialect=excluded.dialect, headers_json=excluded.headers_json,
			insecure=excluded.insecure, proxy=excluded.proxy, updated_at=excluded.updated_at`,
			ep.ID, ep.Name, ep.BaseURL, ep.Dialect, mustJSON(ep.Headers), boolInt(ep.Insecure), ep.Proxy, created.Format(time.RFC3339Nano), ep.UpdatedAt.Format(time.RFC3339Nano))
		if err != nil {
			return err
		}
	}
	ep.KeyHint = keyHint
	ep.HasKey = keyHint != ""
	ep.CreatedAt = created
	return nil
}

// ClearEndpointKey removes the stored key.
func (s *Store) ClearEndpointKey(id string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.db.ExecContext(context.Background(), `UPDATE endpoints SET key_enc=NULL, key_hint='', updated_at=? WHERE id=?`, now(), id)
	return err
}

func maskKey(k string) string {
	if len(k) <= 8 {
		return "••••"
	}
	return k[:4] + "••••" + k[len(k)-4:]
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

const endpointCols = `id,name,base_url,dialect,headers_json,key_enc,key_hint,insecure,proxy,created_at,updated_at`

func scanEndpoint(sc interface{ Scan(...any) error }, withKey bool, codec KeyCodec) (*Endpoint, error) {
	var ep Endpoint
	var headers string
	var keyEnc []byte
	var insecure int
	var created, updated string
	if err := sc.Scan(&ep.ID, &ep.Name, &ep.BaseURL, &ep.Dialect, &headers, &keyEnc, &ep.KeyHint, &insecure, &ep.Proxy, &created, &updated); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(headers), &ep.Headers)
	if ep.Headers == nil {
		ep.Headers = map[string]string{}
	}
	ep.Insecure = insecure == 1
	ep.HasKey = len(keyEnc) > 0
	ep.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	ep.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	if withKey && len(keyEnc) > 0 && codec != nil {
		k, err := codec.Decrypt(keyEnc)
		if err != nil {
			return nil, fmt.Errorf("decrypt key for endpoint %s: %w", ep.ID, err)
		}
		ep.APIKey = k
	}
	return &ep, nil
}

// GetEndpoint returns an endpoint without its key.
func (s *Store) GetEndpoint(id string) (*Endpoint, error) {
	row := s.db.QueryRowContext(context.Background(), `SELECT `+endpointCols+` FROM endpoints WHERE id=?`, id)
	ep, err := scanEndpoint(row, false, s.codec)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return ep, err
}

// GetEndpointWithKey returns an endpoint including its decrypted key.
func (s *Store) GetEndpointWithKey(id string) (*Endpoint, error) {
	row := s.db.QueryRowContext(context.Background(), `SELECT `+endpointCols+` FROM endpoints WHERE id=?`, id)
	ep, err := scanEndpoint(row, true, s.codec)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return ep, err
}

// ListEndpoints returns all endpoints ordered by name.
func (s *Store) ListEndpoints() ([]*Endpoint, error) {
	rows, err := s.db.QueryContext(context.Background(), `SELECT `+endpointCols+` FROM endpoints ORDER BY name, created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Endpoint
	for rows.Next() {
		ep, err := scanEndpoint(rows, false, s.codec)
		if err != nil {
			return nil, err
		}
		out = append(out, ep)
	}
	return out, rows.Err()
}

// DeleteEndpoint removes an endpoint and its overrides (runs are kept).
func (s *Store) DeleteEndpoint(id string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	for _, q := range []string{`DELETE FROM endpoints WHERE id=?`, `DELETE FROM provider_overrides WHERE endpoint_id=?`, `DELETE FROM family_overrides WHERE endpoint_id=?`} {
		if _, err := tx.ExecContext(context.Background(), q, id); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ---------------------------------------------------------------- overrides

// ProviderOverrideRow is a stored provider override.
type ProviderOverrideRow struct {
	RawKey      string `json:"raw_key"`
	DisplayName string `json:"display_name"`
	MergedInto  string `json:"merged_into"`
	Source      string `json:"source"`
}

// ProviderOverrides returns the overrides of an endpoint.
func (s *Store) ProviderOverrides(endpointID string) (map[string]provider.Override, []ProviderOverrideRow, error) {
	rows, err := s.db.QueryContext(context.Background(), `SELECT raw_key, display_name, merged_into, source FROM provider_overrides WHERE endpoint_id=? ORDER BY raw_key`, endpointID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	m := map[string]provider.Override{}
	var list []ProviderOverrideRow
	for rows.Next() {
		var r ProviderOverrideRow
		if err := rows.Scan(&r.RawKey, &r.DisplayName, &r.MergedInto, &r.Source); err != nil {
			return nil, nil, err
		}
		m[r.RawKey] = provider.Override{DisplayName: r.DisplayName, MergedInto: r.MergedInto}
		list = append(list, r)
	}
	return m, list, rows.Err()
}

// SetProviderOverride upserts an override; empty display and merge delete it.
func (s *Store) SetProviderOverride(endpointID string, r ProviderOverrideRow) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if r.DisplayName == "" && r.MergedInto == "" {
		_, err := s.db.ExecContext(context.Background(), `DELETE FROM provider_overrides WHERE endpoint_id=? AND raw_key=?`, endpointID, r.RawKey)
		return err
	}
	if r.Source == "" {
		r.Source = "user"
	}
	_, err := s.db.ExecContext(context.Background(), `INSERT INTO provider_overrides(endpoint_id, raw_key, display_name, merged_into, source, evidence_json, updated_at) VALUES(?,?,?,?,?,'{}',?)
		ON CONFLICT(endpoint_id, raw_key) DO UPDATE SET display_name=excluded.display_name, merged_into=excluded.merged_into, source=excluded.source, updated_at=excluded.updated_at`,
		endpointID, r.RawKey, r.DisplayName, r.MergedInto, r.Source, now())
	return err
}

// FamilyOverrideRow is a stored family override.
type FamilyOverrideRow struct {
	ModelID   string `json:"model_id"`
	FamilyKey string `json:"family_key"`
	Canon     string `json:"canon"`
	Source    string `json:"source"`
}

// FamilyOverrides returns the family overrides of an endpoint.
func (s *Store) FamilyOverrides(endpointID string) (map[string]family.Override, []FamilyOverrideRow, error) {
	rows, err := s.db.QueryContext(context.Background(), `SELECT model_id, family_key, canon, source FROM family_overrides WHERE endpoint_id=? ORDER BY model_id`, endpointID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	m := map[string]family.Override{}
	var list []FamilyOverrideRow
	for rows.Next() {
		var r FamilyOverrideRow
		if err := rows.Scan(&r.ModelID, &r.FamilyKey, &r.Canon, &r.Source); err != nil {
			return nil, nil, err
		}
		m[r.ModelID] = family.Override{Key: r.FamilyKey, Canon: r.Canon}
		list = append(list, r)
	}
	return m, list, rows.Err()
}

// SetFamilyOverride upserts an override; an empty family key deletes it.
func (s *Store) SetFamilyOverride(endpointID string, r FamilyOverrideRow) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if r.FamilyKey == "" {
		_, err := s.db.ExecContext(context.Background(), `DELETE FROM family_overrides WHERE endpoint_id=? AND model_id=?`, endpointID, r.ModelID)
		return err
	}
	if r.Source == "" {
		r.Source = "user"
	}
	_, err := s.db.ExecContext(context.Background(), `INSERT INTO family_overrides(endpoint_id, model_id, family_key, canon, source, evidence_json, updated_at) VALUES(?,?,?,?,?,'{}',?)
		ON CONFLICT(endpoint_id, model_id) DO UPDATE SET family_key=excluded.family_key, canon=excluded.canon, source=excluded.source, updated_at=excluded.updated_at`,
		endpointID, r.ModelID, r.FamilyKey, r.Canon, r.Source, now())
	return err
}

// ---------------------------------------------------------------- runs

// Run is a stored run header.
type Run struct {
	ID           string          `json:"id"`
	Kind         string          `json:"kind"`
	ParentID     string          `json:"parent_id,omitempty"`
	EndpointID   string          `json:"endpoint_id"`
	EndpointName string          `json:"endpoint_name"`
	BaseURL      string          `json:"base_url"`
	Dialect      string          `json:"dialect"`
	Status       string          `json:"status"`
	Preset       string          `json:"preset"`
	Spec         json.RawMessage `json:"spec"`
	Note         string          `json:"note"`
	CreatedAt    time.Time       `json:"created_at"`
	StartedAt    time.Time       `json:"started_at"`
	FinishedAt   time.Time       `json:"finished_at"`
	Summary      json.RawMessage `json:"summary"`
	Progress     json.RawMessage `json:"progress"`
	Error        string          `json:"error,omitempty"`
}

const runCols = `id,kind,parent_id,endpoint_id,endpoint_name,base_url,dialect,status,preset,spec_json,note,created_at,started_at,finished_at,summary_json,progress_json,error`

func scanRun(sc interface{ Scan(...any) error }) (*Run, error) {
	var r Run
	var spec, summary, progress, created string
	var started, finished sql.NullString
	if err := sc.Scan(&r.ID, &r.Kind, &r.ParentID, &r.EndpointID, &r.EndpointName, &r.BaseURL, &r.Dialect, &r.Status, &r.Preset, &spec, &r.Note, &created, &started, &finished, &summary, &progress, &r.Error); err != nil {
		return nil, err
	}
	r.Spec = json.RawMessage(spec)
	r.Summary = json.RawMessage(summary)
	r.Progress = json.RawMessage(progress)
	r.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	r.StartedAt = parseTime(started)
	r.FinishedAt = parseTime(finished)
	return &r, nil
}

// CreateRun inserts a run header; ID is generated when empty.
func (s *Store) CreateRun(r *Run) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if r.ID == "" {
		r.ID = NewID()
	}
	if r.Status == "" {
		r.Status = core.StatusPending
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	spec := string(r.Spec)
	if spec == "" {
		spec = "{}"
	}
	_, err := s.db.ExecContext(context.Background(), `INSERT INTO runs(`+runCols+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,NULL,NULL,'{}','{}','')`,
		r.ID, r.Kind, r.ParentID, r.EndpointID, r.EndpointName, r.BaseURL, r.Dialect, r.Status, r.Preset, spec, r.Note, r.CreatedAt.Format(time.RFC3339Nano))
	return err
}

// SetRunStatus updates the status, timestamps and error of a run.
func (s *Store) SetRunStatus(id, status, errMsg string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	n := now()
	switch status {
	case core.StatusRunning, core.StatusDiscovering:
		_, err := s.db.ExecContext(context.Background(), `UPDATE runs SET status=?, error=?, started_at=COALESCE(started_at, ?) WHERE id=?`, status, errMsg, n, id)
		return err
	case core.StatusDone, core.StatusCancelled, core.StatusFailed, core.StatusInterrupted:
		_, err := s.db.ExecContext(context.Background(), `UPDATE runs SET status=?, error=?, finished_at=? WHERE id=?`, status, errMsg, n, id)
		return err
	}
	_, err := s.db.ExecContext(context.Background(), `UPDATE runs SET status=?, error=? WHERE id=?`, status, errMsg, id)
	return err
}

// SetRunTimes overrides the start/finish timestamps (used by imported data).
func (s *Store) SetRunTimes(id string, started, finished time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.db.ExecContext(context.Background(), `UPDATE runs SET created_at=?, started_at=?, finished_at=? WHERE id=?`, started.UTC().Format(time.RFC3339Nano), started.UTC().Format(time.RFC3339Nano), finished.UTC().Format(time.RFC3339Nano), id)
	return err
}

// SetRunSummary stores the computed summary document.
func (s *Store) SetRunSummary(id string, summary any) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.db.ExecContext(context.Background(), `UPDATE runs SET summary_json=? WHERE id=?`, mustJSON(summary), id)
	return err
}

// SetRunProgress stores the latest progress snapshot.
func (s *Store) SetRunProgress(id string, progress any) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.db.ExecContext(context.Background(), `UPDATE runs SET progress_json=? WHERE id=?`, mustJSON(progress), id)
	return err
}

// SetRunDialect records the detected dialect.
func (s *Store) SetRunDialect(id, dialect string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.db.ExecContext(context.Background(), `UPDATE runs SET dialect=? WHERE id=?`, dialect, id)
	return err
}

// GetRun returns a run header.
func (s *Store) GetRun(id string) (*Run, error) {
	r, err := scanRun(s.db.QueryRowContext(context.Background(), `SELECT `+runCols+` FROM runs WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return r, err
}

// ListRuns returns runs newest first. kind and parent filter when non-empty.
func (s *Store) ListRuns(kind, parent string, limit, offset int) ([]*Run, int, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var where []string
	var args []any
	if kind != "" {
		where = append(where, "kind=?")
		args = append(args, kind)
	}
	if parent != "" {
		where = append(where, "parent_id=?")
		args = append(args, parent)
	} else if kind == "" {
		where = append(where, "parent_id=''")
	}
	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	var total int
	if err := s.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM runs`+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(context.Background(), `SELECT `+runCols+` FROM runs`+clause+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*Run
	for rows.Next() {
		r, err := scanRun(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

// ChildRuns returns the child runs of a suite in creation order.
func (s *Store) ChildRuns(parent string) ([]*Run, error) {
	rows, err := s.db.QueryContext(context.Background(), `SELECT `+runCols+` FROM runs WHERE parent_id=? ORDER BY created_at`, parent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Run
	for rows.Next() {
		r, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// DeleteRun removes a run, its children and all measurements.
func (s *Store) DeleteRun(id string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	ids := []string{id}
	rows, err := s.db.QueryContext(context.Background(), `SELECT id FROM runs WHERE parent_id=?`, id)
	if err != nil {
		return err
	}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			_ = rows.Close()
			return err
		}
		ids = append(ids, c)
	}
	_ = rows.Close()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	for _, rid := range ids {
		for _, tbl := range []string{"runs", "run_models", "attempts", "latency_samples", "load_requests", "cap_results", "quality_results"} {
			col := "run_id"
			if tbl == "runs" {
				col = "id"
			}
			if _, err := tx.ExecContext(context.Background(), `DELETE FROM `+tbl+` WHERE `+col+`=?`, rid); err != nil { //nolint:gosec // table/column names come from a fixed list
				_ = tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}

// MarkInterrupted flags runs that were in progress when the process died.
func (s *Store) MarkInterrupted() (int64, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	res, err := s.db.ExecContext(context.Background(), `UPDATE runs SET status=?, finished_at=?, error='process restarted while the run was in progress' WHERE status IN (?,?,?)`,
		core.StatusInterrupted, now(), core.StatusPending, core.StatusDiscovering, core.StatusRunning)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ---------------------------------------------------------------- run models

// PutRunModels replaces the models of a run.
func (s *Store) PutRunModels(runID string, models []core.RunModel) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(context.Background(), `DELETE FROM run_models WHERE run_id=?`, runID); err != nil {
		_ = tx.Rollback()
		return err
	}
	stmt, err := tx.PrepareContext(context.Background(), `INSERT INTO run_models(run_id,model_id,owned_by,upstream,context_length,max_output,provider_key,provider_name,family_key,family_canon,family_source,family_evidence,raw_json)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, m := range models {
		if _, err := stmt.ExecContext(context.Background(), runID, m.ID, m.OwnedBy, m.Upstream, m.ContextLength, m.MaxOutput, m.ProviderKey, m.ProviderName, m.FamilyKey, m.FamilyCanon, m.FamilySource, m.FamilyEvidence, string(m.Raw)); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// UpdateRunModelAssignments updates provider/family columns without touching
// discovery data.
func (s *Store) UpdateRunModelAssignments(runID string, models []core.RunModel) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(context.Background(), `UPDATE run_models SET provider_key=?, provider_name=?, family_key=?, family_canon=?, family_source=?, family_evidence=? WHERE run_id=? AND model_id=?`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, m := range models {
		if _, err := stmt.ExecContext(context.Background(), m.ProviderKey, m.ProviderName, m.FamilyKey, m.FamilyCanon, m.FamilySource, m.FamilyEvidence, runID, m.ID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ListRunModels returns the models of a run.
func (s *Store) ListRunModels(runID string) ([]core.RunModel, error) {
	rows, err := s.db.QueryContext(context.Background(), `SELECT model_id,owned_by,upstream,context_length,max_output,provider_key,provider_name,family_key,family_canon,family_source,family_evidence,raw_json FROM run_models WHERE run_id=? ORDER BY model_id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.RunModel
	for rows.Next() {
		var m core.RunModel
		var raw string
		if err := rows.Scan(&m.ID, &m.OwnedBy, &m.Upstream, &m.ContextLength, &m.MaxOutput, &m.ProviderKey, &m.ProviderName, &m.FamilyKey, &m.FamilyCanon, &m.FamilySource, &m.FamilyEvidence, &raw); err != nil {
			return nil, err
		}
		if raw != "" {
			m.Raw = json.RawMessage(raw)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------- attempts

// Batch collects rows and writes them in transactions when the buffer is
// full or the flush interval elapsed. It is safe for concurrent use.
type Batch[T any] struct {
	mu       sync.Mutex
	buf      []T
	max      int
	interval time.Duration
	flush    func([]T) error
	timer    *time.Timer
	err      error
}

// NewBatch creates a batch writer.
func NewBatch[T any](max int, interval time.Duration, flush func([]T) error) *Batch[T] {
	return &Batch[T]{max: max, interval: interval, flush: flush}
}

// Add queues a row.
func (b *Batch[T]) Add(v T) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.err != nil {
		return b.err
	}
	b.buf = append(b.buf, v)
	if len(b.buf) >= b.max {
		return b.flushLocked()
	}
	if b.timer == nil {
		b.timer = time.AfterFunc(b.interval, func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			_ = b.flushLocked()
		})
	}
	return nil
}

func (b *Batch[T]) flushLocked() error {
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	if len(b.buf) == 0 {
		return nil
	}
	rows := b.buf
	b.buf = nil
	if err := b.flush(rows); err != nil {
		b.err = err
		return err
	}
	return nil
}

// Flush writes any buffered rows.
func (b *Batch[T]) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked()
}

// AttemptWriter returns a batched writer for attempts of a run.
func (s *Store) AttemptWriter(runID string) *Batch[core.Attempt] {
	return NewBatch(50, 200*time.Millisecond, func(rows []core.Attempt) error { return s.InsertAttempts(runID, rows) })
}

// InsertAttempts writes attempts in one transaction.
func (s *Store) InsertAttempts(runID string, rows []core.Attempt) error {
	if len(rows) == 0 {
		return nil
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(context.Background(), `INSERT INTO attempts(run_id,model_id,round,at_ms,http,inner,conn,latency_ms,ok,class,reason,message,served_model,finish_reason,prompt_tokens,completion_tokens)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, a := range rows {
		if _, err := stmt.ExecContext(context.Background(), runID, a.ModelID, a.Round, a.At.UnixMilli(), a.HTTP, a.Inner, a.Conn, float64(a.Latency)/1e6, boolInt(a.OK), a.Class, a.Reason, a.Message, a.ServedModel, a.FinishReason, a.PromptTokens, a.CompletionTokens); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ListAttempts returns the attempts of a run ordered by round and time.
func (s *Store) ListAttempts(runID string) ([]core.Attempt, error) {
	rows, err := s.db.QueryContext(context.Background(), `SELECT model_id,round,at_ms,http,inner,conn,latency_ms,ok,class,reason,message,served_model,finish_reason,prompt_tokens,completion_tokens FROM attempts WHERE run_id=? ORDER BY round, at_ms, id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.Attempt
	for rows.Next() {
		var a core.Attempt
		var atMs int64
		var latMs float64
		var ok int
		if err := rows.Scan(&a.ModelID, &a.Round, &atMs, &a.HTTP, &a.Inner, &a.Conn, &latMs, &ok, &a.Class, &a.Reason, &a.Message, &a.ServedModel, &a.FinishReason, &a.PromptTokens, &a.CompletionTokens); err != nil {
			return nil, err
		}
		a.At = time.UnixMilli(atMs)
		a.Latency = time.Duration(latMs * 1e6)
		a.OK = ok == 1
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------- latency samples

// LatencyWriter returns a batched writer for latency samples.
func (s *Store) LatencyWriter(runID string) *Batch[core.LatencySample] {
	return NewBatch(50, 200*time.Millisecond, func(rows []core.LatencySample) error { return s.InsertLatencySamples(runID, rows) })
}

// InsertLatencySamples writes samples in one transaction.
func (s *Store) InsertLatencySamples(runID string, rows []core.LatencySample) error {
	if len(rows) == 0 {
		return nil
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(context.Background(), `INSERT INTO latency_samples(run_id,model_id,idx,warm,at_ms,ok,reason,message,ttft_ms,ttst_ms,tpot_ms,e2e_ms,tok_per_s,input_tokens,output_tokens,tokens_exact,served_model)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, l := range rows {
		if _, err := stmt.ExecContext(context.Background(), runID, l.ModelID, l.Index, boolInt(l.Warm), l.At.UnixMilli(), boolInt(l.OK), l.Reason, l.Message, ms(l.TTFT), ms(l.TTST), ms(l.TPOT), ms(l.E2E), l.TokPerSec, l.InputTokens, l.OutputTokens, boolInt(l.TokensExact), l.ServedModel); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func ms(d time.Duration) float64     { return float64(d) / 1e6 }
func fromMs(v float64) time.Duration { return time.Duration(v * 1e6) }

// ListLatencySamples returns the samples of a run.
func (s *Store) ListLatencySamples(runID string) ([]core.LatencySample, error) {
	rows, err := s.db.QueryContext(context.Background(), `SELECT model_id,idx,warm,at_ms,ok,reason,message,ttft_ms,ttst_ms,tpot_ms,e2e_ms,tok_per_s,input_tokens,output_tokens,tokens_exact,served_model FROM latency_samples WHERE run_id=? ORDER BY model_id, idx, id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.LatencySample
	for rows.Next() {
		var l core.LatencySample
		var warm, ok, exact int
		var atMs int64
		var ttft, ttst, tpot, e2e float64
		if err := rows.Scan(&l.ModelID, &l.Index, &warm, &atMs, &ok, &l.Reason, &l.Message, &ttft, &ttst, &tpot, &e2e, &l.TokPerSec, &l.InputTokens, &l.OutputTokens, &exact, &l.ServedModel); err != nil {
			return nil, err
		}
		l.Warm, l.OK, l.TokensExact = warm == 1, ok == 1, exact == 1
		l.At = time.UnixMilli(atMs)
		l.TTFT, l.TTST, l.TPOT, l.E2E = fromMs(ttft), fromMs(ttst), fromMs(tpot), fromMs(e2e)
		out = append(out, l)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------- load requests

// LoadWriter returns a batched writer for load requests.
func (s *Store) LoadWriter(runID string) *Batch[core.LoadRequest] {
	return NewBatch(100, 200*time.Millisecond, func(rows []core.LoadRequest) error { return s.InsertLoadRequests(runID, rows) })
}

// InsertLoadRequests writes load requests in one transaction.
func (s *Store) InsertLoadRequests(runID string, rows []core.LoadRequest) error {
	if len(rows) == 0 {
		return nil
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(context.Background(), `INSERT INTO load_requests(run_id,level,idx,at_ms,ok,reason,ttft_ms,e2e_ms,itl_ms,output_tokens,queue_wait_ms) VALUES(?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, r := range rows {
		if _, err := stmt.ExecContext(context.Background(), runID, r.Level, r.Index, r.At.UnixMilli(), boolInt(r.OK), r.Reason, ms(r.TTFT), ms(r.E2E), ms(r.ITL), r.OutputTokens, ms(r.QueueWait)); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ListLoadRequests returns the load requests of a run.
func (s *Store) ListLoadRequests(runID string) ([]core.LoadRequest, error) {
	rows, err := s.db.QueryContext(context.Background(), `SELECT level,idx,at_ms,ok,reason,ttft_ms,e2e_ms,itl_ms,output_tokens,queue_wait_ms FROM load_requests WHERE run_id=? ORDER BY level, idx, id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.LoadRequest
	for rows.Next() {
		var r core.LoadRequest
		var atMs int64
		var ok int
		var ttft, e2e, itl, qw float64
		if err := rows.Scan(&r.Level, &r.Index, &atMs, &ok, &r.Reason, &ttft, &e2e, &itl, &r.OutputTokens, &qw); err != nil {
			return nil, err
		}
		r.At, r.OK = time.UnixMilli(atMs), ok == 1
		r.TTFT, r.E2E, r.ITL, r.QueueWait = fromMs(ttft), fromMs(e2e), fromMs(itl), fromMs(qw)
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------- capability / quality

// PutCapabilityResults upserts capability probe results.
func (s *Store) PutCapabilityResults(runID string, rows []core.CapabilityResult) error {
	if len(rows) == 0 {
		return nil
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(context.Background(), `INSERT INTO cap_results(run_id,model_id,probe,status,evidence,latency_ms) VALUES(?,?,?,?,?,?)
		ON CONFLICT(run_id,model_id,probe) DO UPDATE SET status=excluded.status, evidence=excluded.evidence, latency_ms=excluded.latency_ms`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, r := range rows {
		if _, err := stmt.ExecContext(context.Background(), runID, r.ModelID, r.Probe, r.Status, r.Evidence, ms(r.Latency)); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ListCapabilityResults returns the capability results of a run.
func (s *Store) ListCapabilityResults(runID string) ([]core.CapabilityResult, error) {
	rows, err := s.db.QueryContext(context.Background(), `SELECT model_id,probe,status,evidence,latency_ms FROM cap_results WHERE run_id=? ORDER BY model_id, probe`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.CapabilityResult
	for rows.Next() {
		var r core.CapabilityResult
		var lat float64
		if err := rows.Scan(&r.ModelID, &r.Probe, &r.Status, &r.Evidence, &lat); err != nil {
			return nil, err
		}
		r.Latency = fromMs(lat)
		out = append(out, r)
	}
	return out, rows.Err()
}

// InsertQualityResults writes quality trials.
func (s *Store) InsertQualityResults(runID string, rows []core.QualityResult) error {
	if len(rows) == 0 {
		return nil
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(context.Background(), `INSERT INTO quality_results(run_id,model_id,question_id,trial,pass,score,output,latency_ms,error) VALUES(?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, r := range rows {
		if _, err := stmt.ExecContext(context.Background(), runID, r.ModelID, r.QuestionID, r.Trial, boolInt(r.Pass), r.Score, r.Output, ms(r.Latency), r.Error); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ListQualityResults returns the quality trials of a run.
func (s *Store) ListQualityResults(runID string) ([]core.QualityResult, error) {
	rows, err := s.db.QueryContext(context.Background(), `SELECT model_id,question_id,trial,pass,score,output,latency_ms,error FROM quality_results WHERE run_id=? ORDER BY model_id, question_id, trial`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.QualityResult
	for rows.Next() {
		var r core.QualityResult
		var pass int
		var lat float64
		if err := rows.Scan(&r.ModelID, &r.QuestionID, &r.Trial, &pass, &r.Score, &r.Output, &lat, &r.Error); err != nil {
			return nil, err
		}
		r.Pass = pass == 1
		r.Latency = fromMs(lat)
		out = append(out, r)
	}
	return out, rows.Err()
}

// WithContext is a helper for callers that want to abort long reads.
func (s *Store) WithContext(ctx context.Context) *sql.DB { _ = ctx; return s.db }
