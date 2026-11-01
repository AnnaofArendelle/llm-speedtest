import { test, type Page } from '@playwright/test';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

// Captures the README screenshot gallery against `llm-speedtest serve --demo`.
// Run: LST_BASE_URL=http://127.0.0.1:8790 npm run screenshots
const here = path.dirname(fileURLToPath(import.meta.url));
const outDir = process.env.LST_SHOT_DIR ?? path.resolve(here, '../../docs/screenshots');

async function settle(page: Page) {
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(900); // let count-up / chart animations finish
}

async function shoot(page: Page, name: string, scheme: string, fullPage = false) {
  await settle(page);
  await page.screenshot({ path: path.join(outDir, `${name}-${scheme}.png`), fullPage, animations: 'disabled' });
}

test.describe('screenshots', () => {
  let suiteId = '';
  test.beforeAll(async ({ request, baseURL }) => {
    const res = await request.get(`${baseURL}/api/runs?limit=20`);
    const body = (await res.json()) as { runs: { id: string; kind: string; parent_id?: string }[] };
    suiteId = body.runs.find((r) => r.kind === 'suite' && !r.parent_id)?.id ?? body.runs[0]?.id ?? '';
  });

  test('pages', async ({ page }, info) => {
    const scheme = info.project.name;
    await page.goto('/');
    await shoot(page, 'dashboard', scheme);
    const tabs = [
      'overview',
      'ranking',
      'models',
      'providers',
      'failures',
      'latency',
      'load',
      'capabilities',
      'quality',
    ];
    for (const tab of tabs) {
      await page.goto(`/runs/${suiteId}?tab=${tab}`);
      await shoot(page, `run-${tab}`, scheme);
    }
    await page.goto(`/runs/${suiteId}?tab=ranking`);
    await settle(page);
    await page.locator('.MuiDataGrid-row').first().click();
    await shoot(page, 'family-drawer', scheme);
    await page.goto('/new');
    await shoot(page, 'new-run', scheme);
    await page.goto('/endpoints');
    await shoot(page, 'endpoints', scheme);
  });

  test('mobile', async ({ browser }, info) => {
    const scheme = info.project.name;
    const ctx = await browser.newContext({
      viewport: { width: 390, height: 844 },
      deviceScaleFactor: 2,
      colorScheme: scheme as 'dark' | 'light',
      locale: 'zh-CN',
      baseURL: info.project.use.baseURL,
    });
    const page = await ctx.newPage();
    await page.goto(`/runs/${suiteId}?tab=overview`);
    await shoot(page, 'mobile-overview', scheme);
    await ctx.close();
  });
});
