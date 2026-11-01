import { useMemo } from 'react';
import { alpha, useTheme } from '@mui/material/styles';
import type { EChartsOption } from 'echarts';
import { EChart } from './EChart';
import { MONO } from './common';

export function Gauge({
  value,
  max = 100,
  label,
  unit = '',
  height = 190,
  color,
  digits = 0,
}: {
  value: number;
  max?: number;
  label: string;
  unit?: string;
  height?: number;
  color?: string;
  digits?: number;
}) {
  const theme = useTheme();
  const option = useMemo<EChartsOption>(
    () => ({
      series: [
        {
          type: 'gauge',
          startAngle: 205,
          endAngle: -25,
          min: 0,
          max,
          radius: '100%',
          center: ['50%', '60%'],
          progress: {
            show: true,
            width: 14,
            roundCap: true,
            itemStyle: { color: color ?? theme.palette.primary.main },
          },
          axisLine: { roundCap: true, lineStyle: { width: 14, color: [[1, alpha(theme.palette.text.primary, 0.08)]] } },
          axisTick: { show: false },
          splitLine: { show: false },
          axisLabel: { show: false },
          pointer: { show: false },
          anchor: { show: false },
          title: { show: true, offsetCenter: [0, '34%'], color: theme.palette.text.secondary, fontSize: 12 },
          detail: {
            valueAnimation: true,
            offsetCenter: [0, '-4%'],
            fontSize: 26,
            fontWeight: 700,
            color: theme.palette.text.primary,
            fontFamily: MONO,
            formatter: (v: number) => `${v.toFixed(digits)}${unit}`,
          },
          data: [{ value: Number(Math.min(value, max).toFixed(digits)), name: label }],
        },
      ],
    }),
    [value, max, label, unit, color, theme, digits],
  );
  return <EChart option={option} height={height} />;
}
