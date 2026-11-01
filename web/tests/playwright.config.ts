import { defineConfig } from '@playwright/test';

const baseURL = process.env.LST_BASE_URL ?? 'http://127.0.0.1:8790';

export default defineConfig({
  testDir: '.',
  testMatch: /.*\.spec\.ts/,
  timeout: 90_000,
  retries: 0,
  reporter: [['list']],
  use: {
    baseURL,
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 1,
    locale: 'zh-CN',
    trace: 'off',
  },
  projects: [
    { name: 'dark', use: { colorScheme: 'dark' } },
    { name: 'light', use: { colorScheme: 'light' } },
  ],
});
