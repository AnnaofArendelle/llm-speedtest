import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Runs from './Runs';
import { mockFetch } from '@/test/mockFetch';
import { globalRoutes } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';

describe('Runs', () => {
  afterEach(() => vi.unstubAllGlobals());
  it('lists runs, enables compare after selecting two scans and confirms deletion', async () => {
    const fm = mockFetch([{ path: /^\/api\/runs\/[^/]+$/, method: 'DELETE', body: { ok: true } }, ...globalRoutes()]);
    renderWithProviders(<Runs />);
    await waitFor(() => expect(screen.getAllByRole('row').length).toBeGreaterThan(1));
    const compare = screen.getByRole('button', { name: 'Compare selected' });
    expect(compare).toBeDisabled();
    const boxes = screen.getAllByRole('checkbox');
    expect(boxes.length).toBeGreaterThanOrEqual(2);
    fireEvent.click(boxes[0]);
    fireEvent.click(boxes[1]);
    expect(compare).toBeEnabled();
    const del = screen.getAllByRole('button', { name: 'Delete' })[0];
    fireEvent.click(del);
    expect(await screen.findByText(/Delete run .* and all its data\?/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));
    await waitFor(() => expect(fm.find('DELETE', '/api/runs/').length).toBe(1));
  });
  it('shows the empty state without runs', async () => {
    mockFetch([{ path: /^\/api\/runs/, body: { runs: [], total: 0 } }]);
    renderWithProviders(<Runs />);
    expect(await screen.findByText('No runs yet')).toBeInTheDocument();
  });
});
