import { useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { withToken } from './client';
import type { InflightItem, RunEvent } from './types';

export interface LiveState {
  connected: boolean;
  latest?: RunEvent;
  /** Most recent attempt/sample results, newest first. */
  feed: RunEvent[];
  /** Round / level / log / cooldown lines, oldest first. */
  rounds: RunEvent[];
  /** Requests currently in progress (from the snapshot and start/attempt events). */
  inflight: InflightItem[];
  done: boolean;
  error?: string;
}

const EMPTY: LiveState = { connected: false, feed: [], rounds: [], inflight: [], done: false };

/** The run's own terminal event (child stages of a suite emit their own `done`). */
export function isTerminalDone(ev: RunEvent, kind: string | undefined): boolean {
  return ev.kind === 'done' && (!ev.stage || ev.stage === kind || ev.stage === 'suite');
}

/** Applies one event to the in-flight list. */
export function applyInflight(list: InflightItem[], ev: RunEvent): InflightItem[] {
  const id = ev.attempt?.model_id ?? ev.sample?.model_id;
  switch (ev.kind) {
    case 'start':
      if (!id) return list;
      return [...list.filter((it) => it.id !== id), { id, stage: ev.stage ?? '', round: ev.round, since: ev.at }];
    case 'attempt':
    case 'sample':
      return id ? list.filter((it) => it.id !== id) : list;
    case 'progress': {
      // capability / quality stages report "<model> <probe|question>" when a request finishes
      const m = ev.message?.split(' ')[0];
      return m ? list.filter((it) => it.id !== m) : list;
    }
    case 'done':
    case 'error':
      return ev.stage === 'suite' || !ev.stage ? [] : list;
    default:
      return list;
  }
}

/** Subscribes to the SSE stream of a run while it is running. */
export function useRunEvents(runId: string | undefined, active: boolean, kind?: string, feedSize = 60): LiveState {
  const qc = useQueryClient();
  const [state, setState] = useState<LiveState>(EMPTY);
  const feedRef = useRef<RunEvent[]>([]);
  const roundsRef = useRef<RunEvent[]>([]);
  const inflightRef = useRef<InflightItem[]>([]);
  useEffect(() => {
    if (!runId || !active) return;
    feedRef.current = [];
    roundsRef.current = [];
    inflightRef.current = [];
    const es = new EventSource(withToken(`/api/runs/${runId}/events`));
    let closed = false;
    const push = (ev: RunEvent) => {
      if (ev.kind === 'snapshot') {
        const data = ev.data as { inflight?: InflightItem[] | null } | undefined;
        inflightRef.current = data?.inflight ?? [];
      } else {
        inflightRef.current = applyInflight(inflightRef.current, ev);
      }
      if (ev.kind === 'attempt' || ev.kind === 'sample') {
        feedRef.current = [ev, ...feedRef.current].slice(0, feedSize);
      }
      if (ev.kind === 'round' || ev.kind === 'level' || ev.kind === 'log' || ev.kind === 'cooldown') {
        roundsRef.current = [...roundsRef.current, ev].slice(-80);
      }
      setState((s) => ({
        ...s,
        connected: true,
        latest: ev.kind === 'snapshot' || ev.kind === 'start' || ev.kind === 'log' ? s.latest : ev,
        feed: feedRef.current,
        rounds: roundsRef.current,
        inflight: inflightRef.current,
      }));
    };
    const parse = (e: MessageEvent): RunEvent | undefined => {
      try {
        return JSON.parse(e.data as string) as RunEvent;
      } catch {
        return undefined;
      }
    };
    const handler = (e: MessageEvent) => {
      const ev = parse(e);
      if (ev) push(ev);
    };
    for (const k of ['snapshot', 'start', 'round', 'attempt', 'progress', 'sample', 'level', 'cooldown', 'log']) {
      es.addEventListener(k, handler);
    }
    es.addEventListener('done', (e: MessageEvent) => {
      const ev = parse(e);
      if (ev) push(ev);
      void qc.invalidateQueries({ queryKey: ['run', runId] });
      void qc.invalidateQueries({ queryKey: ['runs'] });
      if (!ev || isTerminalDone(ev, kind)) {
        closed = true;
        es.close();
        setState((s) => ({ ...s, done: true, connected: false, inflight: [] }));
      }
    });
    es.addEventListener('error', (e: MessageEvent) => {
      if (e.data) {
        const ev = parse(e);
        if (ev) setState((s) => ({ ...s, error: ev.message, done: true, inflight: [] }));
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
  }, [runId, active, kind, feedSize, qc]);
  return state;
}
