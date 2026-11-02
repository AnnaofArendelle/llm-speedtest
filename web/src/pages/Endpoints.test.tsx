import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Endpoints from './Endpoints';
import type { Endpoint } from '@/api/types';
import { mockFetch } from '@/test/mockFetch';
import { fixture, globalRoutes } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';

describe('Endpoints', () => {
  afterEach(() => vi.unstubAllGlobals());
  it('renders stored endpoints and runs a check', async () => {
    mockFetch(globalRoutes());
    renderWithProviders(<Endpoints />);
    const ep = fixture<Endpoint[]>('global', 'endpoints')[0];
    expect(await screen.findByText(ep.name)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Check' }));
    expect(await screen.findByText('Reachable')).toBeInTheDocument();
  });
  it('validates the add dialog and posts the parsed form', async () => {
    const created: Endpoint = { ...fixture<Endpoint[]>('global', 'endpoints')[0], id: 'new', name: 'gw' };
    const fm = mockFetch([{ path: '/api/endpoints', method: 'POST', body: created }, ...globalRoutes()]);
    renderWithProviders(<Endpoints />);
    fireEvent.click(await screen.findByRole('button', { name: 'Add endpoint' }));
    await screen.findByRole('dialog');
    const save = screen.getByRole('button', { name: 'Save' });
    expect(save).toBeDisabled();
    fireEvent.change(screen.getByLabelText(/Base URL/), { target: { value: 'https://gw.example.com/v1' } });
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'gw' } });
    fireEvent.change(screen.getByLabelText(/Extra headers/), { target: { value: 'X-Title=speedtest\nbad line' } });
    expect(save).toBeEnabled();
    fireEvent.click(save);
    await waitFor(() => expect(fm.find('POST', '/api/endpoints').length).toBe(1));
    const body = fm.find('POST', '/api/endpoints')[0].body as {
      base_url: string;
      name: string;
      headers: Record<string, string>;
    };
    expect(body.base_url).toBe('https://gw.example.com/v1');
    expect(body.headers).toEqual({ 'X-Title': 'speedtest' });
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
  });
});
