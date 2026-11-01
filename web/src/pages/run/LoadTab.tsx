import { Box, Stack, Typography } from '@mui/material';
import { DataGrid, type GridColDef } from '@mui/x-data-grid';
import { useTranslation } from 'react-i18next';
import { useLoad } from '@/api/hooks';
import type { LoadLevel } from '@/api/types';
import { EmptyState } from '@/components/EmptyState';
import { Section } from '@/components/Section';
import { KpiCard } from '@/components/KpiCard';
import { ScoreBar } from '@/components/badges';
import { useTokens } from '@/lib/tokens';
import { LoadCurve } from '@/charts/LoadCurve';
import { fmtSeconds } from '@/lib/format';
import { gridDefaults, gridLocale, gridSx } from '@/lib/grid';
import type { TabProps } from './types';

export default function LoadTab({ runId }: TabProps) {
  const { t, i18n } = useTranslation();
  const tk = useTokens();
  const q = useLoad(runId);
  const levels = q.data?.summary?.load ?? [];
  const model = q.data?.summary?.model ?? '';
  const peak = levels.reduce((m, l) => Math.max(m, l.out_tok_per_s), 0);
  const bestGoodput = levels.reduce((m, l) => Math.max(m, l.goodput), 0);
  const columns: GridColDef<LoadLevel>[] = [
    {
      field: 'level',
      headerName: t('run.load.level'),
      width: 110,
      renderCell: (p) => (
        <span className="mono">
          {p.row.mode === 'rate' ? `${(p.row.level / 100).toFixed(2)} rps` : `c=${p.row.level}`}
        </span>
      ),
    },
    { field: 'requests', headerName: t('run.load.requests'), width: 90 },
    { field: 'ok', headerName: t('run.load.ok'), width: 80 },
    {
      field: 'errors',
      headerName: t('run.load.errors'),
      width: 80,
      renderCell: (p) => <span style={{ color: p.row.errors ? tk.fail : undefined }}>{p.row.errors}</span>,
    },
    { field: 'rps', headerName: t('run.load.rps'), width: 110, valueFormatter: (v: number) => v.toFixed(2) },
    {
      field: 'out_tok_per_s',
      headerName: t('run.load.outTps'),
      width: 120,
      valueFormatter: (v: number) => v.toFixed(1),
    },
    {
      field: 'goodput',
      headerName: t('run.load.goodput'),
      width: 160,
      renderCell: (p) => (
        <ScoreBar
          value={p.row.goodput * 100}
          digits={0}
          suffix="%"
          color={p.row.goodput >= 0.9 ? tk.tierA : p.row.goodput >= 0.5 ? tk.tierB : tk.fail}
        />
      ),
    },
    {
      field: 'ttft_s',
      headerName: `${t('run.latency.ttft')} p50 / p90`,
      width: 150,
      valueGetter: (_v, row) => row.ttft_s.median,
      renderCell: (p) => (
        <span className="mono">
          {fmtSeconds(p.row.ttft_s.median)} / {fmtSeconds(p.row.ttft_s.p90)}
        </span>
      ),
    },
    {
      field: 'e2e_s',
      headerName: `${t('run.latency.e2e')} p50 / p90`,
      width: 150,
      valueGetter: (_v, row) => row.e2e_s.median,
      renderCell: (p) => (
        <span className="mono">
          {fmtSeconds(p.row.e2e_s.median)} / {fmtSeconds(p.row.e2e_s.p90)}
        </span>
      ),
    },
    {
      field: 'itl_s',
      headerName: 'ITL p50',
      width: 100,
      valueGetter: (_v, row) => row.itl_s.median,
      valueFormatter: (v: number) => (v > 0 ? `${(v * 1000).toFixed(1)} ms` : '–'),
    },
    { field: 'output_tokens', headerName: t('run.latency.output'), width: 110 },
    {
      field: 'reasons',
      headerName: t('common.reason'),
      width: 200,
      sortable: false,
      renderCell: (p) => (
        <span className="mono" style={{ fontSize: 12 }}>
          {Object.entries(p.row.reasons ?? {})
            .map(([k, v]) => `${k}×${v}`)
            .join(' ')}
        </span>
      ),
    },
  ];
  if (q.error) return <EmptyState title={t('common.empty')} description={String(q.error)} />;
  if (q.isSuccess && levels.length === 0) return <EmptyState title={t('common.empty')} />;
  return (
    <Stack spacing={2}>
      <Typography variant="body2" color="text.secondary">
        {t('run.load.intro')}
      </Typography>
      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr 1fr', md: 'repeat(4, 1fr)' }, gap: 2 }}>
        <KpiCard label={t('run.load.model')} value={model || '–'} hint={q.data?.summary?.mode ?? levels[0]?.mode} />
        <KpiCard label={t('run.load.peak')} value={peak} digits={1} suffix="tok/s" />
        <KpiCard label={t('run.load.bestGoodput')} value={bestGoodput * 100} digits={0} suffix="%" color={tk.tierA} />
        <KpiCard
          label={t('run.load.requests')}
          value={levels.reduce((s, l) => s + l.requests, 0)}
          hint={`${levels.reduce((s, l) => s + l.errors, 0)} ${t('run.load.errors')}`}
        />
      </Box>
      <Section title={t('run.load.curve')}>
        <LoadCurve levels={levels} />
      </Section>
      <Box
        sx={{
          bgcolor: 'background.paper',
          borderRadius: 5,
          border: '1px solid',
          borderColor: 'divider',
          overflow: 'hidden',
        }}
      >
        <DataGrid
          {...gridDefaults}
          rows={levels}
          getRowId={(r) => `${r.mode}-${r.level}`}
          columns={columns}
          loading={q.isLoading}
          localeText={gridLocale(i18n.language)}
          autoHeight
          hideFooter
          sx={{ ...gridSx, '& .MuiDataGrid-row:hover': { cursor: 'default' } }}
        />
      </Box>
    </Stack>
  );
}
