import { defineConfig, type Plugin } from 'vitest/config';
import react from '@vitejs/plugin-react';
import { mkdirSync, writeFileSync } from 'node:fs';
import { fileURLToPath, URL } from 'node:url';

// `emptyOutDir` wipes dist/ on every build; the Go toolchain needs the
// directory to exist for `//go:embed all:dist`, so keep a tracked placeholder.
function keepGitkeep(): Plugin {
  return {
    name: 'keep-gitkeep',
    closeBundle() {
      const dir = fileURLToPath(new URL('./dist', import.meta.url));
      mkdirSync(dir, { recursive: true });
      writeFileSync(`${dir}/.gitkeep`, '');
    },
  };
}

export default defineConfig({
  plugins: [react(), keepGitkeep()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
    chunkSizeWarningLimit: 1600,
    rollupOptions: {
      output: {
        manualChunks: {
          react: ['react', 'react-dom', 'react-router'],
          mui: ['@mui/material', '@mui/icons-material', '@mui/x-data-grid'],
          echarts: ['echarts', 'echarts-for-react'],
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://127.0.0.1:8787',
      '/healthz': 'http://127.0.0.1:8787',
    },
  },
  test: {
    environment: 'happy-dom',
    globals: true,
    css: false,
    setupFiles: ['./src/test/setup.tsx'],
    include: ['src/**/*.test.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text-summary', 'lcov'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/test/**', 'src/**/*.test.*', 'src/main.tsx', 'src/vite-env.d.ts', 'src/app/locales/**'],
      thresholds: { lines: 70, functions: 65, statements: 70, branches: 55 },
    },
  },
});
