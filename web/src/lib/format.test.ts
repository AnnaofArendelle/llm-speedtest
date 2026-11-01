import { describe, expect, it } from 'vitest';
import { fmtSeconds, fmtDurationS, fmtPct, fmtContext, fmtRelative, truncate, fmtRate } from './format';

describe('format', () => {
  it('formats seconds', () => {
    expect(fmtSeconds(0)).toBe('–');
    expect(fmtSeconds(0.25)).toBe('250 ms');
    expect(fmtSeconds(1.234)).toBe('1.23 s');
  });
  it('formats durations', () => {
    expect(fmtDurationS(45)).toBe('45s');
    expect(fmtDurationS(125)).toBe('2m 5s');
    expect(fmtDurationS(3700)).toBe('1h 1m');
  });
  it('formats percentages and rates', () => {
    expect(fmtPct(1, 4)).toBe('25.0%');
    expect(fmtPct(1, 0)).toBe('–');
    expect(fmtRate(0.5)).toBe('50%');
  });
  it('formats context sizes and relative time', () => {
    expect(fmtContext(262144)).toBe('262k');
    expect(fmtContext(0)).toBe('–');
    const now = Date.parse('2026-11-01T00:10:00Z');
    expect(fmtRelative('2026-11-01T00:00:00Z', now)).toBe('10m');
    expect(fmtRelative('0001-01-01T00:00:00Z', now)).toBe('–');
  });
  it('truncates', () => {
    expect(truncate('abcdef', 4)).toBe('abc…');
    expect(truncate(undefined, 4)).toBe('');
  });
});
