import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ProgressCard } from './ProgressCard';
import type { Run } from '@/api/types';
import { renderWithProviders } from '@/test/render';

const run: Run = {
  id: 'r1',
  kind: 'suite',
  endpoint_id: 'e1',
  endpoint_name: 'mock',
  base_url: 'http://x/v1',
  dialect: 'generic',
  status: 'running',
  preset: 'quick',
  spec: null,
  note: '',
  created_at: '2026-11-02T00:00:00Z',
  started_at: '2026-11-02T00:00:00Z',
  finished_at: '0001-01-01T00:00:00Z',
  summary: null,
  progress: null,
};
const child = (kind: Run['kind'], status: Run['status']): Run => ({
  ...run,
  id: `${run.id}-${kind}`,
  kind,
  status,
  parent_id: run.id,
});

describe('ProgressCard', () => {
  it('shows the stage stepper, counters and in-flight rows with skip buttons', () => {
    const onSkip = vi.fn();
    renderWithProviders(
      <ProgressCard
        run={run}
        childRuns={[child('scan', 'done'), child('latency', 'running')]}
        progress={{ stage: 'latency', done: 3, total: 10, ok: 3, fail: 0, eta_s: 42 }}
        inflight={[
          { id: 'a/slow', stage: 'latency', round: 2, since: new Date(Date.now() - 65_000).toISOString() },
          { id: 'b/fast', stage: 'latency', since: new Date().toISOString() },
        ]}
        feed={[]}
        logs={[]}
        onSkip={onSkip}
      />,
    );
    expect(screen.getByTestId('stage-scan-done')).toBeInTheDocument();
    expect(screen.getByTestId('stage-latency-running')).toBeInTheDocument();
    expect(screen.getByText('30%')).toBeInTheDocument();
    expect(screen.getByText(/ETA 42s/)).toBeInTheDocument();
    expect(screen.getAllByTestId('inflight-row')).toHaveLength(2);
    expect(screen.getByText(/1m 5s/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Skip a/slow' }));
    expect(onSkip).toHaveBeenCalledWith('a/slow', false);
    fireEvent.click(screen.getByTestId('skip-all'));
    fireEvent.click(screen.getByRole('button', { name: 'Skip all in flight' }));
    expect(onSkip).toHaveBeenCalledWith(undefined, true);
  });
  it('hides skip controls during the load stage and shows an empty in-flight state', () => {
    renderWithProviders(
      <ProgressCard
        run={{ ...run, kind: 'load' }}
        childRuns={[]}
        progress={{ stage: 'load', done: 0, total: 0 }}
        inflight={[]}
        feed={[]}
        logs={[]}
        onSkip={vi.fn()}
      />,
    );
    expect(screen.getByText('No request in flight')).toBeInTheDocument();
    expect(screen.queryByTestId('skip-all')).toBeNull();
    expect(screen.getByText('…')).toBeInTheDocument();
  });
});
