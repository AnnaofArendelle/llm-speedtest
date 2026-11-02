import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { QueryClientProvider } from '@tanstack/react-query';
import { createElement, type ReactNode } from 'react';
import { applyInflight, isTerminalDone, useRunEvents } from './sse';
import type { RunEvent } from './types';
import { FakeEventSource } from '@/test/fakeEventSource';
import { makeQueryClient } from '@/test/render';

const at = '2026-11-02T00:00:00Z';
const ev = (kind: string, extra: Partial<RunEvent> = {}): RunEvent => ({
  kind,
  done: 0,
  total: 0,
  ok: 0,
  fail: 0,
  at,
  ...extra,
});

describe('applyInflight / isTerminalDone', () => {
  it('tracks start → attempt lifecycle and stage completion', () => {
    let list = applyInflight(
      [],
      ev('start', {
        stage: 'scan',
        round: 1,
        attempt: { model_id: 'a/x', round: 1, at, http: 0, inner: 0, latency: 0, ok: false, class: '', reason: '' },
      }),
    );
    expect(list).toEqual([{ id: 'a/x', stage: 'scan', round: 1, since: at }]);
    list = applyInflight(
      list,
      ev('start', {
        stage: 'scan',
        attempt: { model_id: 'a/y', round: 1, at, http: 0, inner: 0, latency: 0, ok: false, class: '', reason: '' },
      }),
    );
    expect(list.map((i) => i.id)).toEqual(['a/x', 'a/y']);
    list = applyInflight(
      list,
      ev('attempt', {
        attempt: {
          model_id: 'a/x',
          round: 1,
          at,
          http: 200,
          inner: 0,
          latency: 1e9,
          ok: true,
          class: 'ok',
          reason: 'ok',
        },
      }),
    );
    expect(list.map((i) => i.id)).toEqual(['a/y']);
    list = applyInflight(list, ev('progress', { message: 'a/y tools' }));
    expect(list).toEqual([]);
    expect(isTerminalDone(ev('done', { stage: 'scan' }), 'scan')).toBe(true);
    expect(isTerminalDone(ev('done', { stage: 'scan' }), 'suite')).toBe(false);
    expect(isTerminalDone(ev('done', { stage: 'suite' }), 'suite')).toBe(true);
    expect(isTerminalDone(ev('done'), 'latency')).toBe(true);
  });
});

describe('useRunEvents', () => {
  beforeEach(() => FakeEventSource.reset());
  afterEach(() => FakeEventSource.reset());
  const wrapper = ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: makeQueryClient() }, children);

  it('restores in-flight requests from the snapshot and follows events until the terminal done', () => {
    const { result } = renderHook(() => useRunEvents('r1', true, 'scan'), { wrapper });
    const es = FakeEventSource.latest();
    expect(es.url).toContain('/api/runs/r1/events');
    act(() => es.open());
    expect(result.current.connected).toBe(true);
    act(() =>
      es.emit('snapshot', ev('snapshot', { data: { inflight: [{ id: 'a/slow', stage: 'scan', since: at }] } })),
    );
    expect(result.current.inflight).toEqual([{ id: 'a/slow', stage: 'scan', since: at }]);
    act(() => es.emit('round', ev('round', { stage: 'scan', round: 1, rounds: 3, message: 'first' })));
    expect(result.current.rounds).toHaveLength(1);
    act(() =>
      es.emit(
        'attempt',
        ev('attempt', {
          stage: 'scan',
          done: 1,
          total: 2,
          ok: 1,
          attempt: {
            model_id: 'a/slow',
            round: 1,
            at,
            http: 200,
            inner: 0,
            latency: 2e9,
            ok: true,
            class: 'ok',
            reason: 'ok',
          },
        }),
      ),
    );
    expect(result.current.inflight).toEqual([]);
    expect(result.current.feed).toHaveLength(1);
    expect(result.current.latest?.done).toBe(1);
    act(() => es.emit('done', ev('done', { stage: 'latency', message: 'child done' })));
    expect(result.current.done).toBe(false);
    act(() => es.emit('done', ev('done', { stage: 'scan', message: 'scan done' })));
    expect(result.current.done).toBe(true);
    expect(es.readyState).toBe(FakeEventSource.CLOSED);
  });

  it('does not connect while inactive and reports a lost connection', () => {
    const { result, rerender } = renderHook(({ active }) => useRunEvents('r2', active, 'suite'), {
      wrapper,
      initialProps: { active: false },
    });
    expect(FakeEventSource.instances).toHaveLength(0);
    rerender({ active: true });
    const es = FakeEventSource.latest();
    act(() => es.open());
    act(() => es.fail());
    expect(result.current.connected).toBe(false);
    expect(result.current.done).toBe(false);
  });
});
