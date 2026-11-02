import type { ReactElement, ReactNode } from 'react';
import { render, type RenderOptions } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { createMemoryRouter, RouterProvider, type RouteObject } from 'react-router';
import { MotionConfig } from 'framer-motion';
import i18n from '@/app/i18n';
import { ThemeModeProvider } from '@/app/mode';
import { buildTheme } from '@/app/theme';

export function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: 0, gcTime: 0, refetchOnWindowFocus: false },
      mutations: { retry: false },
    },
  });
}

export function Providers({ children, queryClient }: { children: ReactNode; queryClient?: QueryClient }) {
  return (
    <QueryClientProvider client={queryClient ?? makeQueryClient()}>
      <ThemeModeProvider>
        <ThemeProvider theme={buildTheme('light')}>
          <CssBaseline />
          <MotionConfig reducedMotion="always">{children}</MotionConfig>
        </ThemeProvider>
      </ThemeModeProvider>
    </QueryClientProvider>
  );
}

export interface RenderWithProvidersOptions extends Omit<RenderOptions, 'wrapper'> {
  /** Initial URL (path + query) for the memory router. */
  route?: string;
  /** Route path pattern the element is mounted at (default: catch-all). */
  path?: string;
  /** Extra routes, e.g. for links the test follows. */
  routes?: RouteObject[];
  queryClient?: QueryClient;
  lang?: 'zh' | 'en';
}

/** Renders a component inside the app's providers and a memory router. */
export function renderWithProviders(ui: ReactElement, opts: RenderWithProvidersOptions = {}) {
  const { route = '/', path = '*', routes = [], queryClient = makeQueryClient(), lang = 'en', ...rest } = opts;
  void i18n.changeLanguage(lang);
  const router = createMemoryRouter([{ path, element: ui }, ...routes], { initialEntries: [route] });
  const result = render(
    <Providers queryClient={queryClient}>
      <RouterProvider router={router} />
    </Providers>,
    rest,
  );
  return { ...result, router, queryClient };
}
