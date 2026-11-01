import { useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';
import type { EChartsOption } from 'echarts';
import { EChart } from './EChart';
import { MONO, useAxisStyle } from './common';
import { tierColor, tokens } from '@/app/theme';
import { fmtSeconds } from '@/lib/format';
import type { FamilyStat } from '@/api/types';

export function FamilyBars({ families, onPick }: { families: FamilyStat[]; onPick?: (key: string) => void }) {
  const theme = useTheme();
  const axis = useAxisStyle();
  const { t } = useTranslation();
  const tk = tokens[theme.palette.mode];
  const rows = useMemo(() => [...families].reverse(), [families]);
  const option = useMemo<EChartsOption>(
    () => ({
      grid: { left: 8, right: 44, top: 4, bottom: 4, containLabel: true },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        formatter: (params) => {
          const p = (Array.isArray(params) ? params[0] : params) as { dataIndex: number };
          const f = rows[p.dataIndex];
          if (!f) return '';
          return `<b>${f.display}</b><br/>${t('common.score')}: ${f.score.toFixed(1)}<br/>${t('run.ranking.stabilityScore')} ${f.s_stability.toFixed(0)} · ${t('run.ranking.breadth')} ${f.s_breadth.toFixed(0)} · ${t('run.ranking.speed')} ${f.s_speed.toFixed(0)}<br/>${t('common.median')} ${fmtSeconds(f.median_s)} · ${t('common.providers')} ${f.n_prov_ok}`;
        },
      },
      xAxis: { type: 'value', max: 100, ...axis },
      yAxis: {
        type: 'category',
        data: rows.map((f) => f.display),
        axisLabel: {
          color: theme.palette.text.primary,
          fontFamily: MONO,
          fontSize: 12,
          width: 170,
          overflow: 'truncate',
        },
        axisTick: { show: false },
        axisLine: { show: false },
      },
      series: [
        {
          type: 'bar',
          barWidth: 14,
          data: rows.map((f) => ({
            value: Number(f.score.toFixed(1)),
            itemStyle: { color: tierColor(tk, f.tier), borderRadius: 7 },
          })),
          label: { show: true, position: 'right', color: theme.palette.text.secondary, fontFamily: MONO, fontSize: 11 },
        },
      ],
    }),
    [rows, axis, theme, tk, t],
  );
  const height = Math.max(140, rows.length * 26 + 20);
  return (
    <EChart
      option={option}
      height={height}
      onClick={
        onPick
          ? (p) => {
              const f = rows[p.dataIndex ?? -1];
              if (f) onPick(f.key);
            }
          : undefined
      }
    />
  );
}
