import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

// MUI multiplies numeric `borderRadius` values in `sx` by theme.shape.borderRadius
// (12px), so `borderRadius: 5` silently becomes 60px. This guard fails the build
// when a numeric radius ≥ 3 sneaks into a component. Use the `radius` tokens
// from `@/app/theme` or explicit pixel strings instead. ECharts option objects
// (charts/) take real pixels and are exempt.

function walk(dir: string, out: string[] = []): string[] {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, entry.name);
    if (entry.isDirectory()) walk(p, out);
    else if (/\.(tsx?|css)$/.test(entry.name) && !entry.name.endsWith('.test.ts') && !entry.name.endsWith('.test.tsx'))
      out.push(p);
  }
  return out;
}

describe('border radius guard', () => {
  it('uses radius tokens or pixel strings instead of MUI multipliers', () => {
    const root = join(dirname(fileURLToPath(import.meta.url)), '..');
    const offenders: string[] = [];
    for (const file of walk(root)) {
      if (file.includes('/charts/') || file.endsWith('/app/theme.ts')) continue;
      const text = readFileSync(file, 'utf8');
      const re = /borderRadius:\s*(\d+)(?!\s*px|\d|\.)/g;
      let m: RegExpExecArray | null;
      while ((m = re.exec(text))) {
        if (Number(m[1]) >= 3 && Number(m[1]) !== 999) {
          const line = text.slice(0, m.index).split('\n').length;
          offenders.push(`${file.replace(root, 'src')}:${line} borderRadius: ${m[1]}`);
        }
      }
    }
    expect(offenders).toEqual([]);
  });
});
