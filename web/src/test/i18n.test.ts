import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import { zh } from '@/app/locales/zh';
import { en } from '@/app/locales/en';

function walk(dir: string, out: string[] = []): string[] {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, entry.name);
    if (entry.isDirectory()) walk(p, out);
    else if (/\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name)) out.push(p);
  }
  return out;
}

function flatten(obj: Record<string, unknown>, prefix = '', out = new Set<string>()): Set<string> {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (v && typeof v === 'object') flatten(v as Record<string, unknown>, key, out);
    else out.add(key);
  }
  return out;
}

const zhKeys = flatten(zh);
const enKeys = flatten(en);

describe('i18n', () => {
  it('both locales define exactly the same keys', () => {
    expect([...zhKeys].filter((k) => !enKeys.has(k))).toEqual([]);
    expect([...enKeys].filter((k) => !zhKeys.has(k))).toEqual([]);
  });

  it('every static t("…") key used in the sources exists', () => {
    const root = join(dirname(fileURLToPath(import.meta.url)), '..');
    const missing: string[] = [];
    for (const file of walk(root)) {
      const text = readFileSync(file, 'utf8');
      for (const m of text.matchAll(/\bt\(\s*'([a-zA-Z0-9_.]+)'/g)) {
        if (!zhKeys.has(m[1])) missing.push(`${file.replace(root, 'src')}: ${m[1]}`);
      }
    }
    expect(missing).toEqual([]);
  });

  it('dynamic key families cover every runtime value', () => {
    const expectKeys = (keys: string[]) => expect(keys.filter((k) => !zhKeys.has(k))).toEqual([]);
    expectKeys(['suite', 'scan', 'latency', 'load', 'capability', 'quality'].map((k) => `kinds.${k}`));
    expectKeys(
      ['pending', 'discovering', 'running', 'done', 'cancelled', 'failed', 'interrupted'].map((k) => `common.${k}`),
    );
    expectKeys(['A', 'B', 'C', 'none'].map((k) => `tier.${k}`));
    expectKeys(['improved', 'regressed', 'same', 'new', 'gone', 'a', 'b'].map((k) => `compare.${k}`));
    expectKeys(
      [
        'overview',
        'ranking',
        'stability',
        'models',
        'providers',
        'failures',
        'families',
        'latency',
        'load',
        'capabilities',
        'quality',
        'export',
      ].map((k) => `run.tabs.${k}`),
    );
    expectKeys(['supported', 'unsupported', 'error', 'skipped'].map((k) => `run.capabilities.${k}`));
    expectKeys(
      ['quick', 'standard', 'deep'].flatMap((p) => [`newRun.profiles.${p}.name`, `newRun.profiles.${p}.desc`]),
    );
    expectKeys(['dashboard', 'newRun', 'endpoints', 'runs', 'compare', 'settings'].map((k) => `nav.${k}`));
    expectKeys(['system', 'light', 'dark'].map((k) => `settings.theme${k[0].toUpperCase()}${k.slice(1)}`));
  });
});
