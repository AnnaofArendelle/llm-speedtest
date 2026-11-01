import type { Run } from '@/api/types';

export function scanCounts(run: Run | undefined): { models: number; ok: number } | undefined {
  const s = run?.summary;
  if (!s) return undefined;
  if (typeof s.models === 'number' && typeof s.ok === 'number') return { models: s.models, ok: s.ok };
  const nested = (s.summary ?? s.scan) as Record<string, unknown> | undefined;
  if (nested && typeof nested.models === 'number' && typeof nested.ok === 'number')
    return { models: nested.models, ok: nested.ok };
  return undefined;
}

export function isLive(status: string | undefined): boolean {
  return status === 'running' || status === 'pending' || status === 'discovering';
}

export function runTitle(run: Run | undefined): string {
  if (!run) return '';
  return run.endpoint_name || run.base_url;
}

export function hasScan(run: Run): boolean {
  return run.kind === 'scan' || run.kind === 'suite';
}
