import { useMemo } from 'react';
import * as echarts from 'echarts/core';
import { BarChart, GaugeChart, LineChart, PieChart, ScatterChart } from 'echarts/charts';
import {
  DataZoomComponent,
  GridComponent,
  LegendComponent,
  MarkLineComponent,
  TitleComponent,
  TooltipComponent,
} from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';
import ReactEChartsCore from 'echarts-for-react/lib/core';
import type { EChartsOption } from 'echarts';
import { useTheme } from '@mui/material/styles';
import { useReducedMotion } from 'framer-motion';
import { chartPalette } from '@/app/theme';
import type { ClickParams } from './common';

echarts.use([
  BarChart,
  LineChart,
  PieChart,
  ScatterChart,
  GaugeChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  MarkLineComponent,
  TitleComponent,
  CanvasRenderer,
]);

export function EChart({
  option,
  height = 280,
  onClick,
}: {
  option: EChartsOption;
  height?: number | string;
  onClick?: (p: ClickParams) => void;
}) {
  const theme = useTheme();
  const reduce = useReducedMotion();
  const mode = theme.palette.mode;
  const merged = useMemo<EChartsOption>(
    () => ({
      color: [...chartPalette[mode]],
      backgroundColor: 'transparent',
      textStyle: { color: theme.palette.text.secondary, fontFamily: theme.typography.fontFamily },
      animation: !reduce,
      animationDuration: 400,
      animationDurationUpdate: 300,
      ...option,
      tooltip: {
        backgroundColor: theme.palette.background.paper,
        borderColor: theme.palette.divider,
        textStyle: { color: theme.palette.text.primary, fontSize: 12 },
        extraCssText: 'border-radius:10px;box-shadow:0 8px 24px rgba(0,0,0,.18);',
        ...((option.tooltip as object | undefined) ?? {}),
      },
    }),
    [option, mode, theme, reduce],
  );
  return (
    <ReactEChartsCore
      echarts={echarts}
      option={merged}
      notMerge
      lazyUpdate
      style={{ height, width: '100%' }}
      onEvents={onClick ? { click: (p: ClickParams) => onClick(p) } : undefined}
    />
  );
}
