import { screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import RunDetail from '@/pages/RunDetail';
import type { Run } from '@/api/types';
import { arr } from '@/lib/safe';
import { isLive } from '@/lib/run';
import { mockFetch } from '@/test/mockFetch';
import { fixture, globalRoutes, runDoc, SCENARIOS, scenarioRoutes } from '@/test/fixtures';
import type { ReasonInfo } from '@/api/types';
import { renderWithProviders } from '@/test/render';
import { FakeEventSource } from '@/test/fakeEventSource';

const SCAN_TABS = ['overview', 'ranking', 'stability', 'models', 'providers', 'failures', 'families'];

function tabsFor(run: Run, children: Run[]): string[] {
  const kinds = new Set([run.kind, ...children.map((c) => c.kind)]);
  const list: string[] = [];
  if (run.kind === 'scan' || run.kind === 'suite') list.push(...SCAN_TABS);
  if (kinds.has('latency')) list.push('latency');
  if (kinds.has('load')) list.push('load');
  if (kinds.has('capability')) list.push('capabilities');
  if (kinds.has('quality')) list.push('quality');
  list.push('export');
  return list;
}

/** Every tab of every recorded run state must render without an error boundary firing. */
describe.each(SCENARIOS)('RunDetail · %s', (scenario) => {
  const doc = runDoc(scenario);
  const tabs = tabsFor(doc.run, arr(doc.children));
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    FakeEventSource.reset();
  });

  it.each(tabs)('renders the %s tab', async (tab) => {
    const fm = mockFetch([...scenarioRoutes(scenario), ...globalRoutes()]);
    const boundaryErrors: unknown[] = [];
    vi.spyOn(console, 'error').mockImplementation((...args: unknown[]) => {
      if (String(args[0]).includes('[ErrorBoundary]')) boundaryErrors.push(args);
    });
    renderWithProviders(<RunDetail />, { route: `/runs/${doc.run.id}?tab=${tab}`, path: '/runs/:id' });
    await waitFor(() => expect(fm.find('GET', `/api/runs/${doc.run.id}`).length).toBeGreaterThan(0));
    // Let the tab's own queries resolve (fixtures answer immediately).
    await waitFor(() => expect(screen.getAllByRole('tab').length).toBeGreaterThan(0));
    await new Promise((r) => setTimeout(r, 60));
    expect(screen.queryByText(/section failed|Something went wrong/)).toBeNull();
    expect(boundaryErrors).toEqual([]);
    if (isLive(doc.run.status) || doc.running) {
      expect(screen.getByTestId('progress-card')).toBeInTheDocument();
    } else {
      expect(screen.queryByTestId('progress-card')).toBeNull();
    }
  });
});

describe('RunDetail · state-specific content', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    FakeEventSource.reset();
  });
  it('shows the partial-scan empty state while discovering instead of crashing', async () => {
    const doc = runDoc('scan-discovering');
    mockFetch([...scenarioRoutes('scan-discovering'), ...globalRoutes()]);
    renderWithProviders(<RunDetail />, { route: `/runs/${doc.run.id}?tab=overview`, path: '/runs/:id' });
    expect(await screen.findByText('Scan in progress — nothing to summarise yet')).toBeInTheDocument();
    expect(screen.getByTestId('progress-card')).toBeInTheDocument();
  });
  it('renders KPIs and the ranking chart for a finished suite', async () => {
    const doc = runDoc('suite');
    mockFetch([...scenarioRoutes('suite'), ...globalRoutes()]);
    renderWithProviders(<RunDetail />, { route: `/runs/${doc.run.id}?tab=overview`, path: '/runs/:id' });
    expect(await screen.findByText('Models tested')).toBeInTheDocument();
    expect((await screen.findAllByTestId('echart')).length).toBeGreaterThanOrEqual(2);
  });
  it('lists skipped models under their own failure reason', async () => {
    const doc = runDoc('scan-failed');
    mockFetch([...scenarioRoutes('scan-failed'), ...globalRoutes()]);
    renderWithProviders(<RunDetail />, { route: `/runs/${doc.run.id}?tab=failures`, path: '/runs/:id' });
    expect(await screen.findByText('Skipped by user')).toBeInTheDocument();
    const unavailable = fixture<ReasonInfo[]>('global', 'reasons').find((r) => r.code === 'model_unavailable');
    expect(screen.getByText(unavailable?.name_en ?? 'model_unavailable')).toBeInTheDocument();
  });
});
