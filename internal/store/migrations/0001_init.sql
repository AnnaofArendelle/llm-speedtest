-- 0001_init: core schema
CREATE TABLE IF NOT EXISTS endpoints (
  id            TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  base_url      TEXT NOT NULL,
  dialect       TEXT NOT NULL DEFAULT '',
  headers_json  TEXT NOT NULL DEFAULT '{}',
  key_enc       BLOB,
  key_hint      TEXT NOT NULL DEFAULT '',
  insecure      INTEGER NOT NULL DEFAULT 0,
  proxy         TEXT NOT NULL DEFAULT '',
  created_at    TEXT NOT NULL,
  updated_at    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS provider_overrides (
  endpoint_id   TEXT NOT NULL,
  raw_key       TEXT NOT NULL,
  display_name  TEXT NOT NULL DEFAULT '',
  merged_into   TEXT NOT NULL DEFAULT '',
  source        TEXT NOT NULL DEFAULT 'user',
  evidence_json TEXT NOT NULL DEFAULT '{}',
  updated_at    TEXT NOT NULL,
  PRIMARY KEY (endpoint_id, raw_key)
);

CREATE TABLE IF NOT EXISTS family_overrides (
  endpoint_id   TEXT NOT NULL,
  model_id      TEXT NOT NULL,
  family_key    TEXT NOT NULL,
  canon         TEXT NOT NULL DEFAULT '',
  source        TEXT NOT NULL DEFAULT 'user',
  evidence_json TEXT NOT NULL DEFAULT '{}',
  updated_at    TEXT NOT NULL,
  PRIMARY KEY (endpoint_id, model_id)
);

CREATE TABLE IF NOT EXISTS runs (
  id            TEXT PRIMARY KEY,
  kind          TEXT NOT NULL,
  parent_id     TEXT NOT NULL DEFAULT '',
  endpoint_id   TEXT NOT NULL DEFAULT '',
  endpoint_name TEXT NOT NULL DEFAULT '',
  base_url      TEXT NOT NULL DEFAULT '',
  dialect       TEXT NOT NULL DEFAULT '',
  status        TEXT NOT NULL,
  preset        TEXT NOT NULL DEFAULT '',
  spec_json     TEXT NOT NULL DEFAULT '{}',
  note          TEXT NOT NULL DEFAULT '',
  created_at    TEXT NOT NULL,
  started_at    TEXT,
  finished_at   TEXT,
  summary_json  TEXT NOT NULL DEFAULT '{}',
  progress_json TEXT NOT NULL DEFAULT '{}',
  error         TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS runs_created ON runs (created_at);
CREATE INDEX IF NOT EXISTS runs_parent ON runs (parent_id);

CREATE TABLE IF NOT EXISTS run_models (
  run_id          TEXT NOT NULL,
  model_id        TEXT NOT NULL,
  owned_by        TEXT NOT NULL DEFAULT '',
  upstream        TEXT NOT NULL DEFAULT '',
  context_length  INTEGER NOT NULL DEFAULT 0,
  max_output      INTEGER NOT NULL DEFAULT 0,
  provider_key    TEXT NOT NULL DEFAULT '',
  provider_name   TEXT NOT NULL DEFAULT '',
  family_key      TEXT NOT NULL DEFAULT '',
  family_canon    TEXT NOT NULL DEFAULT '',
  family_source   TEXT NOT NULL DEFAULT '',
  family_evidence TEXT NOT NULL DEFAULT '',
  raw_json        TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (run_id, model_id)
);

CREATE TABLE IF NOT EXISTS attempts (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id            TEXT NOT NULL,
  model_id          TEXT NOT NULL,
  round             INTEGER NOT NULL,
  at_ms             INTEGER NOT NULL,
  http              INTEGER NOT NULL DEFAULT 0,
  inner             INTEGER NOT NULL DEFAULT 0,
  conn              TEXT NOT NULL DEFAULT '',
  latency_ms        REAL NOT NULL DEFAULT 0,
  ok                INTEGER NOT NULL,
  class             TEXT NOT NULL DEFAULT '',
  reason            TEXT NOT NULL DEFAULT '',
  message           TEXT NOT NULL DEFAULT '',
  served_model      TEXT NOT NULL DEFAULT '',
  finish_reason     TEXT NOT NULL DEFAULT '',
  prompt_tokens     INTEGER NOT NULL DEFAULT 0,
  completion_tokens INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS attempts_run_model ON attempts (run_id, model_id);
CREATE INDEX IF NOT EXISTS attempts_run_round ON attempts (run_id, round);

CREATE TABLE IF NOT EXISTS latency_samples (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id        TEXT NOT NULL,
  model_id      TEXT NOT NULL,
  idx           INTEGER NOT NULL,
  warm          INTEGER NOT NULL DEFAULT 0,
  at_ms         INTEGER NOT NULL,
  ok            INTEGER NOT NULL,
  reason        TEXT NOT NULL DEFAULT '',
  message       TEXT NOT NULL DEFAULT '',
  ttft_ms       REAL NOT NULL DEFAULT 0,
  ttst_ms       REAL NOT NULL DEFAULT 0,
  tpot_ms       REAL NOT NULL DEFAULT 0,
  e2e_ms        REAL NOT NULL DEFAULT 0,
  tok_per_s     REAL NOT NULL DEFAULT 0,
  input_tokens  INTEGER NOT NULL DEFAULT 0,
  output_tokens INTEGER NOT NULL DEFAULT 0,
  tokens_exact  INTEGER NOT NULL DEFAULT 0,
  served_model  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS latency_run ON latency_samples (run_id, model_id);

CREATE TABLE IF NOT EXISTS load_requests (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id        TEXT NOT NULL,
  level         INTEGER NOT NULL,
  idx           INTEGER NOT NULL,
  at_ms         INTEGER NOT NULL,
  ok            INTEGER NOT NULL,
  reason        TEXT NOT NULL DEFAULT '',
  ttft_ms       REAL NOT NULL DEFAULT 0,
  e2e_ms        REAL NOT NULL DEFAULT 0,
  itl_ms        REAL NOT NULL DEFAULT 0,
  output_tokens INTEGER NOT NULL DEFAULT 0,
  queue_wait_ms REAL NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS load_run ON load_requests (run_id, level);

CREATE TABLE IF NOT EXISTS cap_results (
  run_id     TEXT NOT NULL,
  model_id   TEXT NOT NULL,
  probe      TEXT NOT NULL,
  status     TEXT NOT NULL,
  evidence   TEXT NOT NULL DEFAULT '',
  latency_ms REAL NOT NULL DEFAULT 0,
  PRIMARY KEY (run_id, model_id, probe)
);

CREATE TABLE IF NOT EXISTS quality_results (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id      TEXT NOT NULL,
  model_id    TEXT NOT NULL,
  question_id TEXT NOT NULL,
  trial       INTEGER NOT NULL,
  pass        INTEGER NOT NULL,
  score       REAL NOT NULL DEFAULT 0,
  output      TEXT NOT NULL DEFAULT '',
  latency_ms  REAL NOT NULL DEFAULT 0,
  error       TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS quality_run ON quality_results (run_id, model_id);
