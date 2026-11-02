import { useMemo, type ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { RouterProvider } from 'react-router';
import { ThemeModeProvider } from './mode';
import { useThemeMode } from './modeContext';
import { buildTheme } from './theme';
import { router } from './router';
import { SnackbarProvider } from '@/components/SnackbarProvider';

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 10_000, retry: 1, refetchOnWindowFocus: false } },
});

function Themed({ children }: { children: ReactNode }) {
  const { mode } = useThemeMode();
  const theme = useMemo(() => buildTheme(mode), [mode]);
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline enableColorScheme />
      <SnackbarProvider>{children}</SnackbarProvider>
    </ThemeProvider>
  );
}

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeModeProvider>
        <Themed>
          <RouterProvider router={router} />
        </Themed>
      </ThemeModeProvider>
    </QueryClientProvider>
  );
}
