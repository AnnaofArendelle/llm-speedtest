import { useMemo } from 'react';
import { useTheme } from '@mui/material/styles';

export const MONO = "'JetBrains Mono Variable', ui-monospace, Menlo, monospace";

export interface ClickParams {
  name?: string;
  dataIndex?: number;
  seriesName?: string;
  seriesIndex?: number;
  value?: unknown;
}

export function useAxisStyle() {
  const theme = useTheme();
  return useMemo(
    () => ({
      axisLabel: { color: theme.palette.text.secondary },
      axisLine: { lineStyle: { color: theme.palette.divider } },
      axisTick: { show: false },
      splitLine: { lineStyle: { color: theme.palette.divider } },
    }),
    [theme],
  );
}
