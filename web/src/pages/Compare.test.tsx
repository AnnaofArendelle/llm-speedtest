import { screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Compare from './Compare';
import { mockFetch } from '@/test/mockFetch';
import { globalRoutes } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';

describe('Compare', () => {
  afterEach(() => vi.unstubAllGlobals());
  it('asks for two runs when nothing is selected', async () => {
    mockFetch(globalRoutes());
    renderWithProviders(<Compare />, { route: '/compare' });
    expect(await screen.findByText('Pick two scan or suite runs')).toBeInTheDocument();
  });
  it('renders the diff counters and tables for two runs', async () => {
    mockFetch(globalRoutes());
    renderWithProviders(<Compare />, { route: '/compare?a=x&b=y' });
    expect(await screen.findByText('Model changes')).toBeInTheDocument();
    expect(screen.getByText('Family changes')).toBeInTheDocument();
    expect(screen.getAllByText('Improved').length).toBeGreaterThan(0);
  });
});
