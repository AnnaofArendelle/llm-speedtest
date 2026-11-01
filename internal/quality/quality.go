// Package quality runs a small deterministic question set against models to
// detect broken, degraded or mislabelled deployments without a judge model.
package quality

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"gopkg.in/yaml.v3"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/metrics"
)

//go:embed default.yaml
var defaultYAML []byte

// Question is one scored prompt.
type Question struct {
	ID       string  `yaml:"id" json:"id"`
	Category string  `yaml:"category" json:"category"`
	Prompt   string  `yaml:"prompt" json:"prompt"`
	Type     string  `yaml:"type" json:"type"` // number | exact | contains | regex
	Expect   string  `yaml:"expect" json:"expect"`
	Weight   float64 `yaml:"weight" json:"weight"`
	re       *regexp.Regexp
}

// Set is a question set.
type Set struct {
	Questions      []Question `yaml:"questions" json:"questions"`
	IdentityPrompt string     `yaml:"identity_prompt" json:"identity_prompt"`
}

// Default returns the embedded question set.
func Default() (*Set, error) { return Parse(defaultYAML) }

// Load reads a question set from a YAML file.
func Load(path string) (*Set, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(raw)
}

// Parse decodes and validates a question set.
func Parse(raw []byte) (*Set, error) {
	var s Set
	if err := yaml.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	if len(s.Questions) == 0 {
		return nil, fmt.Errorf("question set has no questions")
	}
	seen := map[string]bool{}
	for i := range s.Questions {
		q := &s.Questions[i]
		if q.ID == "" || q.Prompt == "" || q.Expect == "" {
			return nil, fmt.Errorf("question %d is missing id, prompt or expect", i)
		}
		if seen[q.ID] {
			return nil, fmt.Errorf("duplicate question id %s", q.ID)
		}
		seen[q.ID] = true
		if q.Weight <= 0 {
			q.Weight = 1
		}
		switch q.Type {
		case "number", "exact", "contains":
		case "regex":
			re, err := regexp.Compile(q.Expect)
			if err != nil {
				return nil, fmt.Errorf("question %s: bad regex: %w", q.ID, err)
			}
			q.re = re
		default:
			return nil, fmt.Errorf("question %s: unknown type %q", q.ID, q.Type)
		}
	}
	return &s, nil
}

var numberRE = regexp.MustCompile(`-?\d+(?:\.\d+)?`)

// Check scores an output against a question.
func (q *Question) Check(output string) bool {
	out := strings.TrimSpace(output)
	switch q.Type {
	case "number":
		want, err := strconv.ParseFloat(q.Expect, 64)
		if err != nil {
			return false
		}
		// Accept the expected number anywhere, but require that the first
		// number in the output is the answer (guards against "17 × 23 = 391").
		nums := numberRE.FindAllString(strings.ReplaceAll(out, ",", ""), -1)
		for _, n := range nums {
			if v, err := strconv.ParseFloat(n, 64); err == nil && v == want {
				return true
			}
		}
		return false
	case "exact":
		clean := strings.Trim(out, " \t\r\n.!\"'`*")
		return strings.EqualFold(clean, q.Expect)
	case "contains":
		return strings.Contains(strings.ToLower(out), strings.ToLower(q.Expect))
	case "regex":
		if q.re == nil {
			q.re = regexp.MustCompile(q.Expect)
		}
		return q.re.MatchString(out)
	}
	return false
}

// Spec configures a quality run.
type Spec struct {
	Trials      int           `json:"trials" yaml:"trials"`
	Concurrency int           `json:"concurrency" yaml:"concurrency"`
	PerProvider int           `json:"per_provider" yaml:"per_provider"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	MaxTokens   int           `json:"max_tokens" yaml:"max_tokens"`
	Questions   string        `json:"questions,omitempty" yaml:"questions,omitempty"` // path to a YAML set; empty = built-in
}

// DefaultSpec returns the standard quality spec.
func DefaultSpec() Spec {
	return Spec{Trials: 3, Concurrency: 2, PerProvider: 1, Timeout: 60 * time.Second, MaxTokens: 64}
}

// Normalize fills zero values.
func (s *Spec) Normalize() {
	d := DefaultSpec()
	if s.Trials <= 0 {
		s.Trials = d.Trials
	}
	if s.Concurrency <= 0 {
		s.Concurrency = d.Concurrency
	}
	if s.PerProvider <= 0 {
		s.PerProvider = 1
	}
	if s.Timeout <= 0 {
		s.Timeout = d.Timeout
	}
	if s.MaxTokens <= 0 {
		s.MaxTokens = d.MaxTokens
	}
}

// Target is a model to test.
type Target struct {
	ID       string
	Provider string
	Name     string
}

// Sink receives results.
type Sink interface {
	Quality(ctx context.Context, r core.QualityResult) error
}

// SinkFunc adapts a function to Sink.
type SinkFunc func(ctx context.Context, r core.QualityResult) error

// Quality implements Sink.
func (f SinkFunc) Quality(ctx context.Context, r core.QualityResult) error { return f(ctx, r) }

// Runner executes the question set.
type Runner struct {
	Client *apiclient.Client
	Spec   Spec
	Set    *Set
	Sink   Sink
	Report func(core.Event)
	Stage  string
}

// IdentityQuestionID is the id used for the unscored identity prompt.
const IdentityQuestionID = "_identity"

// Run asks every question Trials times for every target.
func (r *Runner) Run(ctx context.Context, targets []Target) ([]core.QualityResult, error) {
	r.Spec.Normalize()
	if r.Stage == "" {
		r.Stage = "quality"
	}
	if r.Set == nil {
		s, err := Default()
		if err != nil {
			return nil, err
		}
		r.Set = s
	}
	perModel := len(r.Set.Questions)*r.Spec.Trials + 1
	total := len(targets) * perModel
	var mu sync.Mutex
	var all []core.QualityResult
	done := 0
	emit := func(ev core.Event) {
		if r.Report != nil {
			ev.At = time.Now()
			ev.Stage = r.Stage
			r.Report(ev)
		}
	}
	emit(core.Event{Kind: core.EventRound, Round: 1, Rounds: 1, Total: total, Message: "quality"})
	global := semaphore.NewWeighted(int64(r.Spec.Concurrency))
	var pmu sync.Mutex
	perProv := map[string]*semaphore.Weighted{}
	zero := 0.0
	g, gctx := errgroup.WithContext(ctx)
	for _, t := range targets {
		t := t
		g.Go(func() error {
			pmu.Lock()
			ps, ok := perProv[t.Provider]
			if !ok {
				ps = semaphore.NewWeighted(int64(r.Spec.PerProvider))
				perProv[t.Provider] = ps
			}
			pmu.Unlock()
			if err := ps.Acquire(gctx, 1); err != nil {
				return err
			}
			defer ps.Release(1)
			if err := global.Acquire(gctx, 1); err != nil {
				return err
			}
			defer global.Release(1)
			record := func(res core.QualityResult) error {
				if r.Sink != nil {
					if err := r.Sink.Quality(gctx, res); err != nil {
						return err
					}
				}
				mu.Lock()
				all = append(all, res)
				done++
				d := done
				mu.Unlock()
				emit(core.Event{Kind: core.EventProgress, Done: d, Total: total, Message: t.ID + " " + res.QuestionID})
				return nil
			}
			ask := func(prompt string, maxTokens int) (apiclient.ChatResult, time.Duration) {
				pctx, cancel := context.WithTimeout(gctx, r.Spec.Timeout)
				defer cancel()
				start := time.Now()
				res := r.Client.Chat(pctx, apiclient.ChatRequest{Model: t.ID, Messages: []apiclient.Message{{Role: "user", Content: prompt}}, MaxTokens: maxTokens, Temperature: &zero})
				return res, time.Since(start)
			}
			// identity (unscored)
			if r.Set.IdentityPrompt != "" {
				res, lat := ask(r.Set.IdentityPrompt, 32)
				if gctx.Err() != nil {
					return gctx.Err()
				}
				qr := core.QualityResult{ModelID: t.ID, QuestionID: IdentityQuestionID, Trial: 1, Latency: lat, Output: strings.TrimSpace(res.Content)}
				if !res.OK() {
					qr.Error = classify.Evaluate(res.Status, res.Body, t.Name).Reason
				}
				if err := record(qr); err != nil {
					return err
				}
				if !res.OK() && !classify.IsTransient(qr.Error) {
					// deterministic failure: skip the rest for this model
					return nil
				}
			}
			for trial := 1; trial <= r.Spec.Trials; trial++ {
				for i := range r.Set.Questions {
					q := &r.Set.Questions[i]
					if gctx.Err() != nil {
						return gctx.Err()
					}
					res, lat := ask(q.Prompt, r.Spec.MaxTokens)
					if gctx.Err() != nil {
						return gctx.Err()
					}
					qr := core.QualityResult{ModelID: t.ID, QuestionID: q.ID, Trial: trial, Latency: lat, Output: trunc(res.Content, 200)}
					if res.OK() {
						qr.Pass = q.Check(res.Content)
						if qr.Pass {
							qr.Score = q.Weight
						}
					} else {
						qr.Error = classify.Evaluate(res.Status, res.Body, t.Name).Reason
					}
					if err := record(qr); err != nil {
						return err
					}
				}
			}
			return nil
		})
	}
	err := g.Wait()
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].ModelID != all[j].ModelID {
			return all[i].ModelID < all[j].ModelID
		}
		if all[i].Trial != all[j].Trial {
			return all[i].Trial < all[j].Trial
		}
		return all[i].QuestionID < all[j].QuestionID
	})
	if err == nil {
		emit(core.Event{Kind: core.EventDone, Done: done, Total: total, Message: "quality complete"})
	}
	return all, err
}

func trunc(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// ModelScore summarises a model's quality.
type ModelScore struct {
	ModelID     string             `json:"model_id"`
	Score       float64            `json:"score"` // 0..100, median over trials of the weighted pass rate
	Trials      int                `json:"trials"`
	TrialScores []float64          `json:"trial_scores"`
	Consistency float64            `json:"consistency"` // share of questions with identical pass/fail across trials
	Questions   map[string]float64 `json:"questions"`   // per question pass rate
	Categories  map[string]float64 `json:"categories"`  // per category pass rate
	Identity    string             `json:"identity,omitempty"`
	Errors      map[string]int     `json:"errors,omitempty"`
	Answered    int                `json:"answered"`
	Rank        int                `json:"rank,omitempty"`
}

// Summarize scores every model and ranks them by score (ties by
// consistency).
func Summarize(results []core.QualityResult, set *Set) []ModelScore {
	if set == nil {
		set, _ = Default()
	}
	weights := map[string]float64{}
	cats := map[string]string{}
	totalW := 0.0
	for _, q := range set.Questions {
		weights[q.ID] = q.Weight
		cats[q.ID] = q.Category
		totalW += q.Weight
	}
	by := map[string][]core.QualityResult{}
	var order []string
	for _, r := range results {
		if _, ok := by[r.ModelID]; !ok {
			order = append(order, r.ModelID)
		}
		by[r.ModelID] = append(by[r.ModelID], r)
	}
	var out []ModelScore
	for _, id := range order {
		rs := by[id]
		ms := ModelScore{ModelID: id, Questions: map[string]float64{}, Categories: map[string]float64{}, Errors: map[string]int{}}
		trials := map[int]float64{}
		passes := map[string][]bool{}
		catPass, catTot := map[string]float64{}, map[string]float64{}
		for _, r := range rs {
			if r.QuestionID == IdentityQuestionID {
				ms.Identity = r.Output
				if r.Error != "" {
					ms.Errors[r.Error]++
				}
				continue
			}
			if r.Error != "" {
				ms.Errors[r.Error]++
			} else {
				ms.Answered++
			}
			if r.Trial > ms.Trials {
				ms.Trials = r.Trial
			}
			if r.Pass {
				trials[r.Trial] += weights[r.QuestionID]
			}
			passes[r.QuestionID] = append(passes[r.QuestionID], r.Pass)
			catTot[cats[r.QuestionID]] += weights[r.QuestionID]
			if r.Pass {
				catPass[cats[r.QuestionID]] += weights[r.QuestionID]
			}
		}
		var ts []float64
		for t := 1; t <= ms.Trials; t++ {
			if totalW > 0 {
				ts = append(ts, 100*trials[t]/totalW)
			}
		}
		ms.TrialScores = ts
		ms.Score = metrics.Round(metrics.Median(ts), 1)
		consistent, nq := 0, 0
		for qid, ps := range passes {
			nq++
			same := true
			for _, p := range ps {
				if p != ps[0] {
					same = false
					break
				}
			}
			if same {
				consistent++
			}
			n := 0
			for _, p := range ps {
				if p {
					n++
				}
			}
			ms.Questions[qid] = float64(n) / float64(len(ps))
		}
		if nq > 0 {
			ms.Consistency = metrics.Round(float64(consistent)/float64(nq), 3)
		}
		for c, tot := range catTot {
			if tot > 0 {
				ms.Categories[c] = metrics.Round(catPass[c]/tot, 3)
			}
		}
		out = append(out, ms)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if (a.Answered > 0) != (b.Answered > 0) {
			return a.Answered > 0
		}
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.Consistency != b.Consistency {
			return a.Consistency > b.Consistency
		}
		return a.ModelID < b.ModelID
	})
	for i := range out {
		if out[i].Answered > 0 {
			out[i].Rank = i + 1
		}
	}
	return out
}
