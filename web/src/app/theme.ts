import { createTheme, type PaletteMode } from '@mui/material/styles';

export const tokens = {
  light: {
    surface: '#FBFCFE',
    low: '#F3F5F9',
    high: '#E9EDF3',
    on: '#1A1C1E',
    variant: '#45474A',
    primary: '#2563EB',
    secondary: '#6366F1',
    tierA: '#16A34A',
    tierB: '#D97706',
    tierC: '#EA580C',
    fail: '#9F1239',
  },
  dark: {
    surface: '#0B0F19',
    low: '#121826',
    high: '#1A2234',
    on: '#E3E6EC',
    variant: '#9AA3B2',
    primary: '#22D3EE',
    secondary: '#3B82F6',
    tierA: '#4ADE80',
    tierB: '#FBBF24',
    tierC: '#FB923C',
    fail: '#FB7185',
  },
} as const;

export type Tokens = { [K in keyof (typeof tokens)['light']]: string };

export function buildTheme(mode: PaletteMode) {
  const t = tokens[mode];
  return createTheme({
    palette: {
      mode,
      primary: { main: t.primary },
      secondary: { main: t.secondary },
      background: { default: t.surface, paper: t.low },
      text: { primary: t.on, secondary: t.variant },
      success: { main: t.tierA },
      warning: { main: t.tierB },
      error: { main: t.fail },
      divider: mode === 'dark' ? 'rgba(154,163,178,0.16)' : 'rgba(26,28,30,0.1)',
    },
    shape: { borderRadius: 12 },
    typography: {
      fontFamily: `'Inter Variable', 'Noto Sans SC', system-ui, sans-serif`,
      h4: { fontWeight: 700, letterSpacing: -0.5 },
      h5: { fontWeight: 700, letterSpacing: -0.3 },
      h6: { fontWeight: 600 },
      subtitle2: { fontWeight: 600 },
      button: { textTransform: 'none', fontWeight: 600 },
    },
    components: {
      MuiCard: {
        styleOverrides: {
          root: {
            borderRadius: 20,
            backgroundImage: 'none',
            boxShadow: mode === 'dark' ? 'none' : '0 1px 2px rgba(16,24,40,0.06), 0 1px 3px rgba(16,24,40,0.1)',
            border: mode === 'dark' ? '1px solid rgba(154,163,178,0.12)' : '1px solid rgba(26,28,30,0.06)',
          },
        },
      },
      MuiPaper: { styleOverrides: { root: { backgroundImage: 'none' } } },
      MuiChip: { styleOverrides: { root: { borderRadius: 999, fontWeight: 600 } } },
      MuiButton: { styleOverrides: { root: { borderRadius: 999, paddingInline: 18 } } },
      MuiOutlinedInput: { styleOverrides: { root: { borderRadius: 12 } } },
      MuiTab: { styleOverrides: { root: { textTransform: 'none', fontWeight: 600, minHeight: 44 } } },
      MuiTooltip: { styleOverrides: { tooltip: { fontSize: 12, borderRadius: 10 } } },
    },
  });
}

export function tierColor(t: Tokens, tier?: string) {
  switch (tier) {
    case 'A':
      return t.tierA;
    case 'B':
      return t.tierB;
    case 'C':
      return t.tierC;
  }
  return t.fail;
}

export function latencyColor(t: Tokens, seconds: number) {
  if (seconds <= 0) return t.variant;
  if (seconds < 1) return t.tierA;
  if (seconds < 3) return '#84CC16';
  if (seconds < 10) return t.tierB;
  if (seconds < 30) return t.tierC;
  return t.fail;
}

export const chartPalette = {
  light: ['#2563EB', '#6366F1', '#0EA5E9', '#16A34A', '#D97706', '#EA580C', '#DB2777', '#7C3AED', '#0891B2', '#65A30D'],
  dark: ['#22D3EE', '#3B82F6', '#818CF8', '#4ADE80', '#FBBF24', '#FB923C', '#F472B6', '#A78BFA', '#67E8F9', '#A3E635'],
} as const;
