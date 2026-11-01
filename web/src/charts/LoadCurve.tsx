import { useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';
import type { EChartsOption } from 'echarts';
import { EChart } from './EChart';
import { useAxisStyle } from './common';
import type { LoadLevel } from '@/api/types';

export function LoadCurve({ levels, height = 300 }: { levels: LoadLevel[]; height?: number }) {
  const theme = useTheme();
  const axis = useAxisStyle();
  const { t } = useTranslation();
  const option = useMemo<EChartsOption>(
    () => ({
      grid: { left: 8, right: 8, top: 36, bottom: 8, containLabel: true },
      legend: { top: 0, textStyle: { color: theme.palette.text.secondary }, icon: 'circle', itemWidth: 10 },
      tooltip: { trigger: 'axis' },
      xAxis: {
        type: 'category',
        data: levels.map((l) => (l.mode === 'rate' ? `${(l.level / 100).toFixed(2)} rps` : `${l.level}`)),
        ...axis,
      },
      yAxis: [
        { type: 'value', name: 'tok/s', nameTextStyle: { color: theme.palette.text.secondary }, ...axis },
        {
          type: 'value',
          name: 's',
          nameTextStyle: { color: theme.palette.text.secondary },
          axisLabel: axis.axisLabel,
          axisLine: axis.axisLine,
          axisTick: axis.axisTick,
          splitLine: { show: false },
        },
      ],
      series: [
        {
          name: t('run.load.outTps'),
          type: 'bar',
          barMaxWidth: 36,
          data: levels.map((l) => Number(l.out_tok_per_s.toFixed(1))),
          itemStyle: { borderRadius: [6, 6, 0, 0] },
        },
        {
          name: `${t('run.latency.ttft')} p50`,
          type: 'line',
          yAxisIndex: 1,
          smooth: true,
          symbolSize: 7,
          data: levels.map((l) => Number(l.ttft_s.median.toFixed(3))),
          lineStyle: { width: 2.5 },
        },
        {
          name: `${t('run.latency.e2e')} p90`,
          type: 'line',
          yAxisIndex: 1,
          smooth: true,
          symbolSize: 7,
          data: levels.map((l) => Number(l.e2e_s.p90.toFixed(3))),
          lineStyle: { width: 2.5, type: 'dashed' },
        },
      ],
    }),
    [levels, axis, theme, t],
  );
  return <EChart option={option} height={height} />;
}
