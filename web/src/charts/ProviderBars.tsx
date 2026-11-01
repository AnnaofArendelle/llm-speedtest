import { useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';
import type { EChartsOption } from 'echarts';
import { EChart } from './EChart';
import { MONO, useAxisStyle } from './common';
import { tokens } from '@/app/theme';
import type { ProviderStat } from '@/api/types';

export function ProviderBars({
  providers,
  limit = 30,
  onPick,
}: {
  providers: ProviderStat[];
  limit?: number;
  onPick?: (key: string) => void;
}) {
  const theme = useTheme();
  const axis = useAxisStyle();
  const { t } = useTranslation();
  const tk = tokens[theme.palette.mode];
  const rows = useMemo(
    () =>
      [...providers]
        .sort((a, b) => b.models - a.models)
        .slice(0, limit)
        .reverse(),
    [providers, limit],
  );
  const option = useMemo<EChartsOption>(
    () => ({
      grid: { left: 8, right: 24, top: 28, bottom: 4, containLabel: true },
      legend: { top: 0, right: 0, textStyle: { color: theme.palette.text.secondary }, icon: 'circle', itemWidth: 10 },
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      xAxis: { type: 'value', minInterval: 1, ...axis },
      yAxis: {
        type: 'category',
        data: rows.map((p) => p.name),
        axisLabel: {
          color: theme.palette.text.primary,
          fontFamily: MONO,
          fontSize: 12,
          width: 140,
          overflow: 'truncate',
        },
        axisTick: { show: false },
        axisLine: { show: false },
      },
      series: [
        {
          name: t('runs.available'),
          type: 'bar',
          stack: 'a',
          barWidth: 12,
          data: rows.map((p) => p.ok),
          itemStyle: { color: tk.tierA, borderRadius: [0, 0, 0, 0] },
        },
        {
          name: t('tier.none'),
          type: 'bar',
          stack: 'a',
          barWidth: 12,
          data: rows.map((p) => p.models - p.ok),
          itemStyle: { color: tk.fail, borderRadius: [0, 6, 6, 0] },
        },
      ],
    }),
    [rows, axis, theme, tk, t],
  );
  return (
    <EChart
      option={option}
      height={Math.max(160, rows.length * 24 + 44)}
      onClick={
        onPick
          ? (p) => {
              const r = rows[p.dataIndex ?? -1];
              if (r) onPick(r.key);
            }
          : undefined
      }
    />
  );
}
