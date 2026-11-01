import { useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { withToken } from './client';
import type { RunEvent } from './types';

export interface LiveState {
  connected: boolean;
  latest?: RunEvent;
  feed: RunEvent[];
  rounds: RunEvent[];
  done: boolean;
  error?: string;
}

/** Subscribes to the SSE stream of a run while it is running. */
export function useRunEvents(runId: string | undefined, active: boolean, feedSize = 60): LiveState {
  const qc = useQueryClient();
  const [state, setState] = useState<LiveState>({ connected: false, feed: [], rounds: [], done: false });
  const feedRef = useRef<RunEvent[]>([]);
  const roundsRef = useRef<RunEvent[]>([]);
  useEffect(() => {
    if (!runId || !active) return;
    feedRef.current = [];
    roundsRef.current = [];
    const es = new EventSource(withToken(`/api/runs/${runId}/events`));
    let closed = false;
    const push = (ev: RunEvent) => {
      if (ev.kind === 'attempt' || ev.kind === 'sample') {
        feedRef.current = [ev, ...feedRef.current].slice(0, feedSize);
      }
      if (ev.kind === 'round' || ev.kind === 'level' || ev.kind === 'log') {
        roundsRef.current = [...roundsRef.current, ev].slice(-50);
      }
      setState((s) => ({
        ...s,
        connected: true,
        latest: ev.kind === 'snapshot' ? s.latest : ev,
        feed: feedRef.current,
        rounds: roundsRef.current,
      }));
    };
    const handler = (e: MessageEvent) => {
      try {
        push(JSON.parse(e.data as string) as RunEvent);
      } catch {
        /* ignore */
      }
    };
    for (const k of ['snapshot', 'round', 'attempt', 'progress', 'sample', 'level', 'cooldown', 'log']) {
      es.addEventListener(k, handler);
    }
    es.addEventListener('done', (e: MessageEvent) => {
      handler(e);
      let ev: RunEvent | undefined;
      try {
        ev = JSON.parse(e.data as string) as RunEvent;
      } catch {
        /* ignore */
      }
      const finished = !ev || ev.stage === 'suite' || ev.stage === undefined || ev.stage === '';
      void qc.invalidateQueries({ queryKey: ['run', runId] });
      void qc.invalidateQueries({ queryKey: ['runs'] });
      if (finished) {
        closed = true;
        es.close();
        setState((s) => ({ ...s, done: true, connected: false }));
      }
    });
    es.addEventListener('error', (e: MessageEvent) => {
      if (e.data) {
        try {
          const ev = JSON.parse(e.data as string) as RunEvent;
          setState((s) => ({ ...s, error: ev.message, done: true }));
        } catch {
          /* ignore */
        }
      }
      if (closed) return;
      if (es.readyState === EventSource.CLOSED) {
        setState((s) => ({ ...s, connected: false }));
        void qc.invalidateQueries({ queryKey: ['run', runId] });
      }
    });
    es.onopen = () => setState((s) => ({ ...s, connected: true }));
    return () => {
      closed = true;
      es.close();
    };
  }, [runId, active, feedSize, qc]);
  return state;
}
