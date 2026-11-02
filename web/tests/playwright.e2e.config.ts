import { defineConfig } from '@playwright/test';

const baseURL = process.env.LST_BASE_URL ?? 'http://127.0.0.1:8790';

export default defineConfig({
  testDir: '.',
  testMatch: /e2e\.spec\.ts/,
  timeout: 120_000,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never', outputFolder: 'playwright-report' }]] : [['list']],
  use: {
    baseURL,
    viewport: { width: 1440, height: 900 },
    locale: 'en-US',
    colorScheme: 'dark',
    trace: process.env.CI ? 'retain-on-failure' : 'off',
    screenshot: 'only-on-failure',
  },
  projects: [{ name: 'chromium', use: { browserName: 'chromium' } }],
});
