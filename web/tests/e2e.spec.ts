import { expect, test, type Page } from '@playwright/test';
import { startMockOpenAI, type MockOpenAI } from './mock-openai';

// End-to-end checks against a real `llm-speedtest serve --demo` instance (LST_BASE_URL).
// They guard the regressions users hit in v0.1: pages crashing on partial data,
// and exercise the live-run flow including the per-model "skip" control.
test.describe.configure({ mode: 'serial' });

const CRASH_MARKERS = ['Unexpected Application Error', 'Something went wrong', 'section failed'];

function watchErrors(page: Page) {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(`pageerror: ${e.message}`));
  page.on('console', (m) => {
    if (m.type() === 'error' && !/favicon|ERR_CONNECTION|net::|Failed to load resource/i.test(m.text()))
      errors.push(`console: ${m.text()}`);
  });
  return errors;
}

async function expectNoCrash(page: Page, errors: string[], where: string) {
  const text = (await page.locator('body').innerText()).slice(0, 20000);
  for (const marker of CRASH_MARKERS) expect(text, `${where}: ${marker}`).not.toContain(marker);
  expect(errors, `${where}: ${errors.join('\n')}`).toEqual([]);
}

interface RunRow {
  id: string;
  kind: string;
  status: string;
  parent_id?: string;
}

test('every page of the demo renders without errors, including failed and cancelled runs', async ({
  page,
  request,
  baseURL,
}) => {
  const errors = watchErrors(page);
  for (const path of ['/', '/runs', '/endpoints', '/new', '/compare', '/settings']) {
    await page.goto(path);
    await page.waitForLoadState('networkidle');
    await expectNoCrash(page, errors, path);
  }
  const res = await request.get(`${baseURL}/api/runs?limit=50`);
  const { runs } = (await res.json()) as { runs: RunRow[] };
  const suite = runs.find((r) => r.kind === 'suite' && !r.parent_id && r.status === 'done');
  expect(suite, 'demo suite run').toBeTruthy();
  const tabs = [
    'overview',
    'ranking',
    'stability',
    'models',
    'providers',
    'failures',
    'families',
    'latency',
    'load',
    'capabilities',
    'quality',
    'export',
  ];
  for (const tab of tabs) {
    await page.goto(`/runs/${suite!.id}?tab=${tab}`);
    await page.waitForLoadState('networkidle');
    await expectNoCrash(page, errors, `suite/${tab}`);
  }
  // Drawers: open the first ranking row and the first model row.
  await page.goto(`/runs/${suite!.id}?tab=ranking`);
  await page.locator('.MuiDataGrid-row').first().click();
  await expect(page.getByText('Capability matrix')).toBeVisible({ timeout: 10_000 });
  await expectNoCrash(page, errors, 'family drawer');
  for (const status of ['failed', 'cancelled']) {
    const run = runs.find((r) => r.status === status && !r.parent_id);
    expect(run, `demo ${status} run`).toBeTruthy();
    for (const tab of ['overview', 'models', 'failures', 'export']) {
      await page.goto(`/runs/${run!.id}?tab=${tab}`);
      await page.waitForLoadState('networkidle');
      await expectNoCrash(page, errors, `${status}/${tab}`);
    }
  }
});

test('a live scan shows progress and a hanging model can be skipped', async ({ page, request, baseURL }) => {
  let mock: MockOpenAI | undefined;
  try {
    mock = await startMockOpenAI();
    const errors = watchErrors(page);
    const created = await request.post(`${baseURL}/api/endpoints`, {
      data: { name: 'e2e-mock', base_url: `${mock.url}/v1` },
    });
    const endpoint = (await created.json()) as { id: string };

    await page.goto(`/new?endpoint=${endpoint.id}`);
    await page.getByText('Quick', { exact: true }).click();
    for (const label of ['Latency lab (top N after the scan)', 'Capability probes', 'Quality spot-check']) {
      await page.getByLabel(label).uncheck();
    }
    await page.getByRole('button', { name: 'Start', exact: true }).click();
    await page.waitForURL(/\/runs\/[^/?]+/);
    const progress = page.getByTestId('progress-card');
    await expect(progress).toBeVisible({ timeout: 15_000 });

    const row = page.getByTestId('inflight-row').filter({ hasText: 'beta/hang' });
    await expect(row).toBeVisible({ timeout: 30_000 });
    await row.getByRole('button', { name: 'Skip beta/hang' }).click();
    // The toast and the progress-card log line both mention the model; assert on the toast.
    await expect(page.getByRole('alert').filter({ hasText: 'Skipped beta/hang' })).toBeVisible({ timeout: 10_000 });

    await expect(page.getByText('Done', { exact: true }).first()).toBeVisible({ timeout: 60_000 });
    await page.waitForLoadState('networkidle');
    await expectNoCrash(page, errors, 'live run');
    expect(mock.hits['beta/hang']).toBe(1);

    await page.goto(`${page.url().split('?')[0]}?tab=failures`);
    await expect(page.getByText('Skipped by user').first()).toBeVisible({ timeout: 10_000 });
    await page.goto(`${page.url().split('?')[0]}?tab=models`);
    await expect(page.getByText('beta/hang')).toBeVisible();
    await expect(page.getByText('alpha/fast')).toBeVisible();
    await expectNoCrash(page, errors, 'after skip');
  } finally {
    await mock?.close();
  }
});

test('a run against an unreachable endpoint ends as failed and its pages still render', async ({
  page,
  request,
  baseURL,
}) => {
  const errors = watchErrors(page);
  const started = await request.post(`${baseURL}/api/runs`, {
    data: { kind: 'scan', endpoint: { base_url: 'http://127.0.0.1:9/v1' }, profile: 'quick' },
  });
  expect(started.status()).toBe(202);
  const { id } = (await started.json()) as { id: string };
  await expect
    .poll(
      async () => {
        const r = await request.get(`${baseURL}/api/runs/${id}`);
        const doc = (await r.json()) as { run: { status: string } };
        return doc.run.status;
      },
      { timeout: 30_000 },
    )
    .toBe('failed');
  for (const tab of ['overview', 'models', 'export']) {
    await page.goto(`/runs/${id}?tab=${tab}`);
    await page.waitForLoadState('networkidle');
    await expectNoCrash(page, errors, `failed/${tab}`);
  }
  await expect(page.getByText('Failed', { exact: true }).first()).toBeVisible();
});
