import { describe, expect, it } from 'vitest';
import type { Run } from '@/api/types';
import { hasScan, isLive, runTitle, scanCounts } from './run';

const base: Run = {
  id: 'r',
  kind: 'scan',
  endpoint_id: '',
  endpoint_name: '',
  base_url: 'http://x/v1',
  dialect: '',
  status: 'done',
  preset: 'quick',
  spec: null,
  note: '',
  created_at: '',
  started_at: '',
  finished_at: '',
  summary: null,
  progress: null,
};

describe('run helpers', () => {
  it('reads scan counts from flat and nested summaries', () => {
    expect(scanCounts(base)).toBeUndefined();
    expect(scanCounts({ ...base, summary: { models: 3, ok: 2 } })).toEqual({ models: 3, ok: 2 });
    expect(scanCounts({ ...base, summary: { summary: { models: 5, ok: 1 } } })).toEqual({ models: 5, ok: 1 });
    expect(scanCounts(undefined)).toBeUndefined();
  });
  it('classifies statuses and titles', () => {
    expect(isLive('running')).toBe(true);
    expect(isLive('done')).toBe(false);
    expect(hasScan({ ...base, kind: 'suite' })).toBe(true);
    expect(hasScan({ ...base, kind: 'latency' })).toBe(false);
    expect(runTitle(base)).toBe('http://x/v1');
    expect(runTitle({ ...base, endpoint_name: 'gw' })).toBe('gw');
    expect(runTitle(undefined)).toBe('');
  });
});
