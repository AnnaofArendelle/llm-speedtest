import { screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Dashboard from './Dashboard';
import { mockFetch } from '@/test/mockFetch';
import { globalRoutes, scenarioRoutes } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';

describe('Dashboard', () => {
  afterEach(() => vi.unstubAllGlobals());
  it('shows the empty state when there are no runs', async () => {
    mockFetch([
      { path: /^\/api\/runs/, body: { runs: [], total: 0 } },
      { path: /^\/api\/endpoints$/, body: [] },
      ...globalRoutes().filter((r) => !/runs|endpoints\$/.test(String(r.path))),
    ]);
    renderWithProviders(<Dashboard />);
    expect(await screen.findByText('No runs yet. Add an endpoint, then start a speed test.')).toBeInTheDocument();
  });
  it('renders the latest run headline, charts, recent runs and endpoints', async () => {
    mockFetch([...scenarioRoutes('suite'), ...globalRoutes()]);
    renderWithProviders(<Dashboard />);
    expect(await screen.findByText('Models tested')).toBeInTheDocument();
    expect(await screen.findByText('Recent runs')).toBeInTheDocument();
    expect((await screen.findAllByTestId('echart')).length).toBeGreaterThanOrEqual(2);
    expect(screen.getAllByText('Test now').length).toBeGreaterThan(0);
  });
});
