import { useMemo } from 'react';
import { alpha, useTheme } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';
import type { EChartsOption } from 'echarts';
import type { ProfileDimension } from '@/api/types';
import { EChart } from './EChart';

/** Radar of the 0–100 profile dimensions; uncovered axes are greyed out. */
export function RadarChart({ dims, height = 260, name }: { dims: ProfileDimension[]; height?: number; name?: string }) {
  const theme = useTheme();
  const { i18n } = useTranslation();
  const zh = i18n.language.startsWith('zh');
  const option = useMemo<EChartsOption>(() => {
    const primary = theme.palette.primary.main;
    return {
      tooltip: {
        trigger: 'item',
        formatter: () =>
          `<b>${name ?? ''}</b><br/>` +
          dims
            .map(
              (d) =>
                `${zh ? d.label_zh : d.label_en}: ${d.covered ? `<b>${d.score.toFixed(0)}</b> · ${d.detail}` : '–'}`,
            )
            .join('<br/>'),
      },
      radar: {
        radius: '56%',
        center: ['50%', '52%'],
        splitNumber: 4,
        indicator: dims.map((d) => ({
          name: zh ? d.label_zh : d.label_en,
          max: 100,
          color: d.covered ? theme.palette.text.secondary : theme.palette.text.disabled,
        })),
        axisName: { fontSize: 10, overflow: 'truncate', width: 64 },
        splitLine: { lineStyle: { color: theme.palette.divider } },
        axisLine: { lineStyle: { color: theme.palette.divider } },
        splitArea: { areaStyle: { color: [alpha(primary, 0.05), 'transparent'] } },
      },
      series: [
        {
          type: 'radar',
          symbolSize: 5,
          data: [
            {
              name: name ?? '',
              value: dims.map((d) => (d.covered ? Math.round(d.score) : 0)),
              areaStyle: { color: alpha(primary, 0.22) },
              lineStyle: { color: primary, width: 2 },
              itemStyle: { color: primary },
            },
          ],
        },
      ],
    };
  }, [dims, theme, zh, name]);
  return <EChart option={option} height={height} />;
}
