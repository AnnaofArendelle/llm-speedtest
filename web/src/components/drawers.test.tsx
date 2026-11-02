import { screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ModelDrawer } from './ModelDrawer';
import { FamilyDrawer } from './FamilyDrawer';
import { fixture, globalRoutes, scenarioRoutes } from '@/test/fixtures';
import { mockFetch } from '@/test/mockFetch';
import { renderWithProviders } from '@/test/render';

describe('drawers', () => {
  afterEach(() => vi.unstubAllGlobals());
  it('model drawer shows stats, the capability matrix and the attempt history', async () => {
    mockFetch([...scenarioRoutes('suite'), ...globalRoutes()]);
    const { model } = fixture<{ model: { id: string } }>('suite', 'model');
    renderWithProviders(<ModelDrawer runId="r" modelId={model.id} onClose={() => undefined} />);
    expect(await screen.findByText('Attempts per round')).toBeInTheDocument();
    expect(await screen.findByText('Capability matrix')).toBeInTheDocument();
    expect(screen.getAllByText(model.id).length).toBeGreaterThan(0);
  });
  it('family drawer shows members, providers and the capability matrix', async () => {
    mockFetch([...scenarioRoutes('suite'), ...globalRoutes()]);
    const { family } = fixture<{ family: { key: string; display: string } }>('suite', 'family');
    renderWithProviders(<FamilyDrawer runId="r" familyKey={family.key} onClose={() => undefined} />);
    expect(await screen.findByText(/Members \(\d+\)/)).toBeInTheDocument();
    expect(await screen.findByText('Capability matrix')).toBeInTheDocument();
  });
});
