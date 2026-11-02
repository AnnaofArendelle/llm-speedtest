import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ModelsTab } from './ModelsTab';
import { ProvidersTab } from './ProvidersTab';
import ExportTab from './ExportTab';
import type { Run } from '@/api/types';
import { mockFetch } from '@/test/mockFetch';
import { fixture, globalRoutes, runDoc, scenarioRoutes } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';

const noop = () => undefined;

describe('run tabs', () => {
  afterEach(() => vi.unstubAllGlobals());
  const run: Run = { ...runDoc('suite').run, endpoint_id: 'e1' };

  it('ModelsTab filters drive the server query and rows open the model drawer', async () => {
    const fm = mockFetch([...scenarioRoutes('suite'), ...globalRoutes()]);
    const onModel = vi.fn();
    renderWithProviders(<ModelsTab runId={run.id} run={run} onModel={onModel} onFamily={noop} goTab={noop} />, {
      route: `/runs/${run.id}?tab=models`,
      path: '/runs/:id',
    });
    const first = fixture<{ models: { id: string }[] }>('suite', 'models').models[0];
    expect(await screen.findByText(first.id)).toBeInTheDocument();
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0]); // status filter
    fireEvent.click(await screen.findByRole('option', { name: 'Available' }));
    await waitFor(() =>
      expect(fm.calls.some((c) => c.path.endsWith('/models') && c.url.searchParams.get('status') === 'ok')).toBe(true),
    );
    fireEvent.change(screen.getByLabelText('Search'), { target: { value: 'fast' } });
    await waitFor(
      () =>
        expect(fm.calls.some((c) => c.path.endsWith('/models') && c.url.searchParams.get('q') === 'fast')).toBe(true),
      { timeout: 2000 },
    );
    fireEvent.click(screen.getByText(first.id));
    expect(onModel).toHaveBeenCalledWith(first.id);
  });

  it('ProvidersTab renames a provider through the override API', async () => {
    const fm = mockFetch([
      { path: /\/api\/endpoints\/e1\/overrides\/provider/, method: 'PUT', body: { ok: true } },
      ...scenarioRoutes('suite'),
      ...globalRoutes(),
    ]);
    const goTab = vi.fn();
    renderWithProviders(<ProvidersTab runId={run.id} run={run} onModel={noop} onFamily={noop} goTab={goTab} />);
    const edit = (await screen.findAllByRole('button', { name: 'Edit' }))[0];
    fireEvent.click(edit);
    const dialog = await screen.findByRole('dialog');
    fireEvent.change(screen.getByLabelText('Display name'), { target: { value: 'Alpha Cloud' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => expect(fm.find('PUT', '/api/endpoints/e1/overrides/provider')).toHaveLength(1));
    expect(
      (fm.find('PUT', '/api/endpoints/e1/overrides/provider')[0].body as { display_name: string }).display_name,
    ).toBe('Alpha Cloud');
    expect(dialog).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    fireEvent.click(screen.getAllByRole('button', { name: 'Models' })[0]);
    expect(goTab).toHaveBeenCalledWith('models', expect.objectContaining({ provider: expect.any(String) }));
  });

  it('ExportTab offers downloads in the chosen language and previews the Markdown', async () => {
    mockFetch([...scenarioRoutes('suite'), ...globalRoutes()]);
    renderWithProviders(<ExportTab runId={run.id} run={run} onModel={noop} onFamily={noop} goTab={noop} />);
    const md = screen.getByRole('link', { name: 'Markdown report' });
    expect(md.getAttribute('href')).toContain(`/api/runs/${run.id}/report.md?lang=en&download=1`);
    expect(screen.getByRole('link', { name: 'JSON data' }).getAttribute('href')).toContain('/export.json');
    fireEvent.click(screen.getByRole('button', { name: 'Preview' }));
    expect(await screen.findByText(/LLM API Speed Test Report|#/)).toBeInTheDocument();
  });
});
