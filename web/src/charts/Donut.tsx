import { useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import type { EChartsOption } from 'echarts';
import { EChart } from './EChart';

export interface DonutDatum {
  key: string;
  name: string;
  value: number;
  color?: string;
}

export function Donut({
  data,
  height = 260,
  onPick,
}: {
  data: DonutDatum[];
  height?: number;
  onPick?: (key: string) => void;
}) {
  const theme = useTheme();
  const option = useMemo<EChartsOption>(
    () => ({
      tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
      legend: {
        orient: 'vertical',
        right: 0,
        top: 'middle',
        type: 'scroll',
        icon: 'circle',
        itemWidth: 10,
        textStyle: { color: theme.palette.text.secondary, fontSize: 12, overflow: 'truncate', width: 150 },
        pageIconColor: theme.palette.text.secondary,
        pageTextStyle: { color: theme.palette.text.secondary },
      },
      series: [
        {
          type: 'pie',
          radius: ['54%', '80%'],
          center: ['30%', '50%'],
          avoidLabelOverlap: true,
          itemStyle: { borderRadius: 6, borderColor: theme.palette.background.paper, borderWidth: 2 },
          label: { show: false },
          emphasis: { scale: true, scaleSize: 5 },
          data: data.map((d) => ({
            name: d.name,
            value: d.value,
            itemStyle: d.color ? { color: d.color } : undefined,
          })),
        },
      ],
    }),
    [data, theme],
  );
  return (
    <EChart
      option={option}
      height={height}
      onClick={
        onPick
          ? (p) => {
              const d = data[p.dataIndex ?? -1];
              if (d) onPick(d.key);
            }
          : undefined
      }
    />
  );
}
