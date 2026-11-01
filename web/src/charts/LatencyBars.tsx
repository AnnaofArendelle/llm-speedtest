import { useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';
import type { EChartsOption } from 'echarts';
import { EChart } from './EChart';
import { MONO, useAxisStyle } from './common';
import type { LatencyModelStats } from '@/api/types';

export function LatencyBars({ stats, height = 320 }: { stats: LatencyModelStats[]; height?: number }) {
  const theme = useTheme();
  const axis = useAxisStyle();
  const { t } = useTranslation();
  const rows = useMemo(() => stats.filter((s) => s.ok > 0), [stats]);
  const option = useMemo<EChartsOption>(
    () => ({
      grid: { left: 8, right: 8, top: 36, bottom: 8, containLabel: true },
      legend: { top: 0, textStyle: { color: theme.palette.text.secondary }, icon: 'circle', itemWidth: 10 },
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      xAxis: {
        type: 'category',
        data: rows.map((s) => s.model_id),
        axisLabel: {
          ...axis.axisLabel,
          fontFamily: MONO,
          fontSize: 10,
          rotate: rows.length > 6 ? 28 : 0,
          width: 120,
          overflow: 'truncate',
        },
        axisLine: axis.axisLine,
        axisTick: axis.axisTick,
      },
      yAxis: [
        { type: 'value', name: 's', nameTextStyle: { color: theme.palette.text.secondary }, ...axis },
        {
          type: 'value',
          name: 'tok/s',
          nameTextStyle: { color: theme.palette.text.secondary },
          axisLabel: axis.axisLabel,
          axisLine: axis.axisLine,
          axisTick: axis.axisTick,
          splitLine: { show: false },
        },
      ],
      series: [
        {
          name: `${t('run.latency.ttft')} p50`,
          type: 'bar',
          barMaxWidth: 26,
          data: rows.map((s) => Number(s.ttft_s.median.toFixed(3))),
          itemStyle: { borderRadius: [6, 6, 0, 0] },
        },
        {
          name: `${t('run.latency.e2e')} p50`,
          type: 'bar',
          barMaxWidth: 26,
          data: rows.map((s) => Number(s.e2e_s.median.toFixed(3))),
          itemStyle: { borderRadius: [6, 6, 0, 0] },
        },
        {
          name: `${t('run.latency.tps')} p50`,
          type: 'line',
          yAxisIndex: 1,
          smooth: true,
          symbolSize: 7,
          data: rows.map((s) => Number(s.tok_per_s.median.toFixed(1))),
          lineStyle: { width: 2.5 },
        },
      ],
    }),
    [rows, axis, theme, t],
  );
  return <EChart option={option} height={height} />;
}
