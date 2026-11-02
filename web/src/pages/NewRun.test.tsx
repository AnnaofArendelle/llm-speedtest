import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import NewRun from './NewRun';
import type { Endpoint } from '@/api/types';
import { mockFetch } from '@/test/mockFetch';
import { fixture, globalRoutes } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';

describe('NewRun', () => {
  afterEach(() => vi.unstubAllGlobals());
  const ep = fixture<Endpoint[]>('global', 'endpoints')[0];

  it('requires an endpoint before starting', async () => {
    mockFetch(globalRoutes());
    renderWithProviders(<NewRun />, { route: '/new' });
    expect(await screen.findByText('Select or enter an endpoint first')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Start' })).toBeDisabled();
  });

  it('starts a suite for the preselected endpoint and demands confirmation for load tests', async () => {
    const fm = mockFetch([
      { path: '/api/runs', method: 'POST', status: 202, body: { id: 'run-1' } },
      ...globalRoutes(),
    ]);
    renderWithProviders(<NewRun />, { route: `/new?endpoint=${ep.id}` });
    const start = await screen.findByRole('button', { name: 'Start' });
    await waitFor(() => expect(start).toBeEnabled());
    fireEvent.click(screen.getByLabelText(/Load test on the best model/));
    expect(start).toBeDisabled();
    fireEvent.click(screen.getByLabelText(/I understand a load test/));
    expect(start).toBeEnabled();
    fireEvent.click(start);
    await waitFor(() => expect(fm.find('POST', '/api/runs').length).toBe(1));
    const body = fm.find('POST', '/api/runs')[0].body as {
      kind: string;
      profile: string;
      endpoint_id: string;
      confirm: boolean;
      suite: { scan: boolean; load: boolean; filter: { exclude_owners: string[] | null } };
    };
    expect(body.kind).toBe('suite');
    expect(body.profile).toBe('standard');
    expect(body.endpoint_id).toBe(ep.id);
    expect(body.confirm).toBe(true);
    expect(body.suite.scan).toBe(true);
    expect(body.suite.load).toBe(true);
    expect(body.suite.filter.exclude_owners).toEqual(['combo']);
  });
});
