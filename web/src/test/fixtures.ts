import type { Run } from '@/api/types';
import type { MockRoute } from './mockFetch';

// Fixtures are real API responses recorded by the Go contract test:
//   go test ./internal/server -run TestAPIContract -update-fixtures
const json = import.meta.glob('./fixtures/**/*.json', { eager: true, import: 'default' }) as Record<string, unknown>;
const text = import.meta.glob('./fixtures/**/*.md', { eager: true, query: '?raw', import: 'default' }) as Record<
  string,
  string
>;

export const SCENARIOS = [
  'suite',
  'scan-discovering',
  'scan-running',
  'scan-failed',
  'scan-all-fail',
  'latency-only',
  'load-only',
] as const;
export type Scenario = (typeof SCENARIOS)[number];

export function hasFixture(scenario: string, name: string): boolean {
  return `./fixtures/${scenario}/${name}.json` in json;
}

export function fixture<T = unknown>(scenario: string, name: string): T {
  const v = json[`./fixtures/${scenario}/${name}.json`];
  if (v === undefined) throw new Error(`missing fixture ${scenario}/${name}`);
  return structuredClone(v) as T;
}

export function runDoc(scenario: string) {
  return fixture<{ run: Run; children: Run[] | null; running: boolean }>(scenario, 'run');
}

function jsonRoute(path: RegExp, scenario: string, name: string): MockRoute {
  const body = json[`./fixtures/${scenario}/${name}.json`];
  return body === undefined ? { path, status: 404, body: { error: `no fixture ${scenario}/${name}` } } : { path, body };
}

/** Routes serving one scenario's run endpoints (any run id). */
export function scenarioRoutes(scenario: string): MockRoute[] {
  const run = '/api/runs/[^/?]+';
  const report = text[`./fixtures/${scenario}/report.md`];
  return [
    jsonRoute(new RegExp(`^${run}/summary`), scenario, 'summary'),
    jsonRoute(new RegExp(`^${run}/families\\?order=stability`), scenario, 'families-stability'),
    jsonRoute(new RegExp(`^${run}/families(\\?|$)`), scenario, 'families-composite'),
    jsonRoute(new RegExp(`^${run}/families/`), scenario, 'family'),
    jsonRoute(new RegExp(`^${run}/providers`), scenario, 'providers'),
    jsonRoute(new RegExp(`^${run}/models(\\?|$)`), scenario, 'models'),
    jsonRoute(new RegExp(`^${run}/models/`), scenario, 'model'),
    jsonRoute(new RegExp(`^${run}/failures`), scenario, 'failures'),
    jsonRoute(new RegExp(`^${run}/latency`), scenario, 'latency'),
    jsonRoute(new RegExp(`^${run}/load`), scenario, 'load'),
    jsonRoute(new RegExp(`^${run}/capabilities`), scenario, 'capabilities'),
    jsonRoute(new RegExp(`^${run}/quality`), scenario, 'quality'),
    jsonRoute(new RegExp(`^${run}/profile/model/`), scenario, 'profile-model'),
    jsonRoute(new RegExp(`^${run}/profile/family/`), scenario, 'profile-family'),
    jsonRoute(new RegExp(`^${run}/export\\.json`), scenario, 'export'),
    { path: new RegExp(`^${run}/report\\.md`), body: report ?? '# report', contentType: 'text/markdown' },
    { path: new RegExp(`^${run}/skip`), method: 'POST', body: { skipped: ['a/slow'] } },
    { path: new RegExp(`^${run}/cancel`), method: 'POST', body: { cancelled: true } },
    jsonRoute(new RegExp(`^${run}$`), scenario, 'run'),
  ];
}

/** Routes for the global endpoints (reasons, presets, endpoints, run list, …). */
export function globalRoutes(): MockRoute[] {
  const g = (name: string, re: RegExp): MockRoute => jsonRoute(re, 'global', name);
  return [
    g('version', /^\/api\/version/),
    g('reasons', /^\/api\/reasons/),
    g('probes', /^\/api\/probes/),
    g('presets', /^\/api\/presets/),
    g('dialects', /^\/api\/dialects/),
    g('endpoints', /^\/api\/endpoints$/),
    g('compare', /^\/api\/compare/),
    g('runs', /^\/api\/runs(\?|$)/),
    { path: /^\/api\/endpoints\/[^/]+\/check/, method: 'POST', body: json['./fixtures/global/check.json'] },
    { path: /^\/api\/endpoints\/[^/]+\/discover/, method: 'POST', body: json['./fixtures/global/discover.json'] },
    { path: /^\/api\/check/, method: 'POST', body: json['./fixtures/global/check.json'] },
    { path: /^\/api\/discover/, method: 'POST', body: json['./fixtures/global/discover.json'] },
    { path: /^\/api\/endpoints\/[^/]+\/overrides$/, body: { providers: [], families: [] } },
  ];
}
