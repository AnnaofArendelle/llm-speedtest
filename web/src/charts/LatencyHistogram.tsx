import { useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import type { EChartsOption } from 'echarts';
import { EChart } from './EChart';
import { useAxisStyle } from './common';
import { latencyColor, tokens } from '@/app/theme';
import type { Bucket } from '@/api/types';

export function LatencyHistogram({
  buckets,
  height = 250,
  onPick,
}: {
  buckets: Bucket[];
  height?: number;
  onPick?: (b: Bucket) => void;
}) {
  const theme = useTheme();
  const axis = useAxisStyle();
  const tk = tokens[theme.palette.mode];
  const option = useMemo<EChartsOption>(
    () => ({
      grid: { left: 8, right: 8, top: 28, bottom: 4, containLabel: true },
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      xAxis: { type: 'category', data: buckets.map((b) => b.label), ...axis },
      yAxis: { type: 'value', minInterval: 1, ...axis },
      series: [
        {
          type: 'bar',
          barMaxWidth: 52,
          data: buckets.map((b) => ({
            value: b.count,
            itemStyle: { color: latencyColor(tk, b.hi > 0 ? (b.lo + b.hi) / 2 : b.lo + 1), borderRadius: [8, 8, 0, 0] },
          })),
          label: { show: true, position: 'top', color: theme.palette.text.secondary, fontSize: 11 },
        },
      ],
    }),
    [buckets, axis, tk, theme],
  );
  return (
    <EChart
      option={option}
      height={height}
      onClick={
        onPick
          ? (p) => {
              const b = buckets[p.dataIndex ?? -1];
              if (b) onPick(b);
            }
          : undefined
      }
    />
  );
}
