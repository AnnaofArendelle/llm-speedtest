import { screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { LiveFeed } from './LiveFeed';
import type { RunEvent } from '@/api/types';
import { mockFetch } from '@/test/mockFetch';
import { globalRoutes } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';

const at = '2026-11-02T00:00:00Z';
const base = { done: 0, total: 0, ok: 0, fail: 0, at };

describe('LiveFeed', () => {
  afterEach(() => vi.unstubAllGlobals());
  it('renders attempts and samples with latency or reason chips', async () => {
    mockFetch(globalRoutes());
    const events: RunEvent[] = [
      {
        ...base,
        kind: 'attempt',
        seq: 1,
        attempt: {
          model_id: 'a/fast',
          round: 1,
          at,
          http: 200,
          inner: 0,
          latency: 1.2e9,
          ok: true,
          class: 'ok',
          reason: 'ok',
        },
      },
      {
        ...base,
        kind: 'attempt',
        seq: 2,
        attempt: {
          model_id: 'b/slow',
          round: 1,
          at,
          http: 0,
          inner: 0,
          latency: 60e9,
          ok: false,
          class: 'timeout',
          reason: 'timeout',
        },
      },
      {
        ...base,
        kind: 'sample',
        seq: 3,
        sample: {
          model_id: 'a/fast',
          index: 0,
          warm: true,
          at,
          ok: true,
          ttft: 0.42e9,
          ttst: 0,
          tpot: 0,
          e2e: 2e9,
          tok_per_s: 55,
          input_tokens: 10,
          output_tokens: 20,
          tokens_exact: true,
        },
      },
    ];
    renderWithProviders(<LiveFeed events={events} />);
    expect(screen.getByText('1.20 s')).toBeInTheDocument();
    expect(await screen.findByText('Timeout')).toBeInTheDocument();
    expect(screen.getByText(/TTFT 420 ms · 55.0 tok\/s/)).toBeInTheDocument();
    expect(screen.getAllByText('a/fast')).toHaveLength(2);
  });
  it('shows a waiting message without events', () => {
    renderWithProviders(<LiveFeed events={[]} />);
    expect(screen.getByText('Waiting for events…')).toBeInTheDocument();
  });
});
