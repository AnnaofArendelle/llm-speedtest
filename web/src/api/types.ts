export interface Run {
  id: string;
  kind: 'suite' | 'scan' | 'latency' | 'load' | 'capability' | 'quality';
  parent_id?: string;
  endpoint_id: string;
  endpoint_name: string;
  base_url: string;
  dialect: string;
  status: 'pending' | 'discovering' | 'running' | 'done' | 'cancelled' | 'failed' | 'interrupted';
  preset: string;
  spec: unknown;
  note: string;
  created_at: string;
  started_at: string;
  finished_at: string;
  summary: Record<string, unknown> | null;
  progress: Progress | null;
  error?: string;
  running?: boolean;
}

export interface Progress {
  stage?: string;
  round?: number;
  rounds?: number;
  done?: number;
  total?: number;
  ok?: number;
  fail?: number;
  eta_s?: number;
  at?: string;
}

export interface Endpoint {
  id: string;
  name: string;
  base_url: string;
  dialect: string;
  headers: Record<string, string>;
  key_hint: string;
  has_key: boolean;
  insecure: boolean;
  proxy: string;
  created_at: string;
  updated_at: string;
}

export interface DialectInfo {
  name: string;
  version?: string;
  latest?: string;
  healthy?: boolean;
  evidence?: string[];
  score?: number;
}

export interface CheckResult {
  base_url: string;
  reachable: boolean;
  auth_ok: boolean;
  models: number;
  owners: number;
  dialect: DialectInfo;
  latency_s: number;
  error?: string;
  suggested_base_url?: string;
}

export interface RunModelRow {
  id: string;
  owned_by: string;
  upstream: string;
  provider_key: string;
  provider_name: string;
  family_key: string;
  family_canon: string;
  family_source: string;
}

export interface ProviderInfo {
  key: string;
  name: string;
  models: number;
  conns?: string[];
  raw_keys?: string[];
}

export interface DiscoverResult {
  models: RunModelRow[] | null;
  total: number;
  filtered: number;
  providers: ProviderInfo[] | null;
  provider_field: string;
  provider_rationale: string;
  families: number;
  dialect: DialectInfo;
}

export interface Bucket {
  label: string;
  lo: number;
  hi: number;
  count: number;
}

export interface Summary {
  models: number;
  ok: number;
  fail: number;
  provider_level_fail: number;
  families: number;
  families_ok: number;
  families_unavailable: number;
  providers: number;
  tier_counts: Record<string, number> | null;
  family_tier_counts: Record<string, number> | null;
  latency_median_s: number;
  latency_p90_s: number;
  latency_min_s: number;
  latency_max_s: number;
  buckets: Bucket[] | null;
  reasons: Record<string, number> | null;
  attempts: number;
  rounds: number;
}

export interface Scoring {
  w_stability: number;
  w_breadth: number;
  w_speed: number;
  speed_lo_s: number;
  speed_hi_s: number;
}

export interface Attempt {
  model_id: string;
  round: number;
  at: string;
  http: number;
  inner: number;
  conn?: string;
  latency: number; // nanoseconds
  ok: boolean;
  class: string;
  reason: string;
  message?: string;
  served_model?: string;
  finish_reason?: string;
  prompt_tokens?: number;
  completion_tokens?: number;
}

export interface ModelStat {
  id: string;
  owned_by: string;
  upstream: string;
  provider_key: string;
  provider_name: string;
  family_key: string;
  family_canon: string;
  family_source: string;
  family_evidence?: string;
  context_length?: number;
  status: 'ok' | 'fail' | 'untested';
  attempts: number;
  ok: number;
  rate: number;
  tier?: string;
  min_s: number;
  median_s: number;
  max_s: number;
  jitter: number;
  served_model?: string;
  flaky?: Record<string, number>;
  reason: string;
  first_reason?: string;
  http?: number;
  inner?: number;
  message?: string;
  fail_latency_s?: number;
  /** Attempts cancelled by the user through "skip" (excluded from the statistics). */
  skipped?: number;
}

export interface FamilyStat {
  key: string;
  display: string;
  member_ids: string[] | null;
  n: number;
  ok: number;
  providers: Record<string, number> | null;
  ok_providers: Record<string, number> | null;
  n_prov_ok: number;
  n_prov_stable: number;
  attempts: number;
  successes: number;
  rate: number;
  tier?: string;
  min_s: number;
  median_s: number;
  max_s: number;
  best_id?: string;
  stable_ids: string[] | null;
  score: number;
  s_stability: number;
  s_breadth: number;
  s_speed: number;
  reasons: Record<string, number> | null;
  sources: Record<string, number> | null;
  rank_composite?: number;
  rank_stability?: number;
}

export interface ProviderStat {
  key: string;
  name: string;
  models: number;
  families: number;
  ok: number;
  rate: number;
  median_s: number;
  top_reason?: string;
  top_reason_count?: number;
  reasons: Record<string, number> | null;
  account_level_all: boolean;
  evidence?: { field: string; value: string }[];
}

export interface ReasonInfo {
  code: string;
  name: string;
  name_en: string;
  description: string;
  description_en: string;
  fix: string;
  fix_en: string;
  provider_level: boolean;
  transient: boolean;
}

export interface RunSummaryDoc {
  run: Run;
  /** Null while a run is still discovering models or has no stored summary. */
  summary: Summary | null;
  scoring?: Scoring;
  provider_field?: string;
  provider_rationale?: string;
  provider_suggestions?: { keys: string[]; evidence: string }[] | null;
  top_families?: FamilyStat[] | null;
  top_models?: ModelStat[] | null;
  providers?: ProviderStat[] | null;
  /** True while the run is in progress or did not finish; data may be incomplete. */
  partial?: boolean;
  status?: string;
  kind?: string;
}

export interface Stats {
  count: number;
  min: number;
  max: number;
  mean: number;
  median: number;
  p75: number;
  p90: number;
  p95: number;
  p99: number;
  stddev: number;
}

export interface LatencyModelStats {
  model_id: string;
  samples: number;
  ok: number;
  cold?: { ttft_s: number; e2e_s: number; tok_per_s: number };
  ttft_s: Stats;
  ttst_s: Stats;
  tpot_s: Stats;
  e2e_s: Stats;
  tok_per_s: Stats;
  output_tokens: Stats;
  input_tok_per_s: number;
  tokens_exact: boolean;
  served_model?: string;
  reasons?: Record<string, number>;
  rank?: number;
}

export interface LatencySample {
  model_id: string;
  index: number;
  warm: boolean;
  at: string;
  ok: boolean;
  reason?: string;
  ttft: number;
  ttst: number;
  tpot: number;
  e2e: number;
  tok_per_s: number;
  input_tokens: number;
  output_tokens: number;
  tokens_exact: boolean;
}

export interface LoadLevel {
  level: number;
  mode: string;
  requests: number;
  ok: number;
  errors: number;
  duration: number;
  rps: number;
  out_tok_per_s: number;
  goodput: number;
  ttft_s: Stats;
  e2e_s: Stats;
  itl_s: Stats;
  queue_wait_s: Stats;
  output_tokens: number;
  reasons?: Record<string, number>;
}

export interface CapMatrix {
  models: string[] | null;
  probes: string[] | null;
  cells: Record<string, Record<string, { status: string; evidence?: string; latency_s?: number }> | null> | null;
  support: Record<string, number> | null;
}

export interface QualityScore {
  model_id: string;
  score: number;
  trials: number;
  trial_scores: number[] | null;
  consistency: number;
  questions: Record<string, number> | null;
  categories: Record<string, number> | null;
  identity?: string;
  errors?: Record<string, number>;
  answered: number;
  rank?: number;
}

export interface ProbeInfo {
  id: string;
  zh: string;
  en: string;
}

export interface RunEvent {
  seq?: number;
  kind: string;
  run_id?: string;
  stage?: string;
  round?: number;
  rounds?: number;
  done: number;
  total: number;
  ok: number;
  fail: number;
  eta_s?: number;
  message?: string;
  attempt?: Attempt;
  sample?: LatencySample;
  data?: unknown;
  at: string;
}

export interface FailureGroup {
  reason: string;
  info: ReasonInfo;
  count: number;
  by_provider: Record<string, ModelStat[] | null> | null;
}

export interface CompareDoc {
  a: Run;
  b: Run;
  summary_a: Summary;
  summary_b: Summary;
  counts: Record<string, number> | null;
  models:
    | {
        id: string;
        provider: string;
        family: string;
        status_a: string;
        status_b: string;
        tier_a?: string;
        tier_b?: string;
        median_a: number;
        median_b: number;
        delta_median: number;
        reason_a?: string;
        reason_b?: string;
        change: string;
      }[]
    | null;
  families:
    | {
        key: string;
        display: string;
        score_a: number;
        score_b: number;
        rank_a: number;
        rank_b: number;
        median_a: number;
        median_b: number;
        ok_a: number;
        ok_b: number;
      }[]
    | null;
}

export interface ProfileDimension {
  key: string;
  label_zh: string;
  label_en: string;
  score: number;
  covered: boolean;
  detail: string;
}

export interface Profile {
  id: string;
  kind: 'model' | 'family';
  display: string;
  dimensions: ProfileDimension[] | null;
  coverage: Record<string, boolean> | null;
  measured: Record<string, number> | null;
  members: number;
}

export interface InflightItem {
  id: string;
  stage: string;
  round?: number;
  since: string;
}

export interface SkipResponse {
  skipped: string[] | null;
}

export interface SuiteSpec {
  profile: string;
  scan: boolean;
  latency: boolean;
  latency_top: number;
  capability: boolean;
  capability_top: number;
  quality: boolean;
  quality_top: number;
  load: boolean;
  load_model: string;
  filter: {
    exclude_owners: string[] | null;
    include_owners: string[] | null;
    id_regex: string;
    exclude_regex: string;
    limit: number;
    models: string[] | null;
  };
  specs: Record<string, unknown>;
  lang: string;
}
