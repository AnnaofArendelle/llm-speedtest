import { useMemo } from 'react';
import { Box, Chip, Stack, Typography } from '@mui/material';
import { DataGrid, type GridColDef } from '@mui/x-data-grid';
import { useTranslation } from 'react-i18next';
import { radius } from '@/app/theme';
import { useLatency } from '@/api/hooks';
import type { LatencyModelStats } from '@/api/types';
import { EmptyState } from '@/components/EmptyState';
import { LoadingBlock } from '@/components/LoadingBlock';
import { Section } from '@/components/Section';
import { useTokens } from '@/lib/tokens';
import { LatencyBars } from '@/charts/LatencyBars';
import { Gauge } from '@/charts/Gauge';
import { fmtSeconds } from '@/lib/format';
import { gridDefaults, gridLocale, gridSx } from '@/lib/grid';
import type { TabProps } from './types';

export default function LatencyTab({ runId, onModel }: TabProps) {
  const { t, i18n } = useTranslation();
  const tk = useTokens();
  const q = useLatency(runId);
  const stats = useMemo(() => (q.data?.stats ?? []).filter((s) => s.ok > 0), [q.data]);
  const best = useMemo(() => {
    if (!stats.length) return undefined;
    const byTtft = [...stats].sort((a, b) => a.ttft_s.median - b.ttft_s.median)[0];
    const byTps = [...stats].sort((a, b) => b.tok_per_s.median - a.tok_per_s.median)[0];
    const byE2e = [...stats].sort((a, b) => a.e2e_s.median - b.e2e_s.median)[0];
    return {
      byTtft,
      byTps,
      byE2e,
      maxTps: Math.max(...stats.map((s) => s.tok_per_s.max)),
      maxTtft: Math.max(...stats.map((s) => s.ttft_s.median)),
      maxE2e: Math.max(...stats.map((s) => s.e2e_s.median)),
    };
  }, [stats]);
  const columns: GridColDef<LatencyModelStats>[] = [
    { field: 'rank', headerName: '#', width: 60 },
    {
      field: 'model_id',
      headerName: t('common.model'),
      flex: 1,
      minWidth: 240,
      renderCell: (p) => (
        <Typography variant="body2" className="mono" noWrap title={p.row.model_id}>
          {p.row.model_id}
        </Typography>
      ),
    },
    {
      field: 'ok',
      headerName: t('common.samples'),
      width: 90,
      renderCell: (p) => (
        <span className="mono">
          {p.row.ok}/{p.row.samples}
        </span>
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
      field: 'ttst_s',
      headerName: `${t('run.latency.ttst')} p50`,
      width: 110,
      valueGetter: (_v, row) => row.ttst_s.median,
      valueFormatter: (v: number) => fmtSeconds(v),
    },
    {
      field: 'tpot_s',
      headerName: `${t('run.latency.tpot')} p50`,
      width: 110,
      valueGetter: (_v, row) => row.tpot_s.median,
      valueFormatter: (v: number) => (v > 0 ? `${(v * 1000).toFixed(1)} ms` : '–'),
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
      field: 'tok_per_s',
      headerName: `${t('run.latency.tps')} p50 (max)`,
      width: 150,
      valueGetter: (_v, row) => row.tok_per_s.median,
      renderCell: (p) => (
        <span className="mono" style={{ fontWeight: 700 }}>
          {p.row.tok_per_s.median.toFixed(1)}{' '}
          <span style={{ opacity: 0.6, fontWeight: 400 }}>({p.row.tok_per_s.max.toFixed(1)})</span>
        </span>
      ),
    },
    {
      field: 'output_tokens',
      headerName: t('run.latency.output'),
      width: 110,
      valueGetter: (_v, row) => row.output_tokens.mean,
      valueFormatter: (v: number) => v.toFixed(0),
    },
    {
      field: 'input_tok_per_s',
      headerName: t('run.latency.inputTps'),
      width: 120,
      valueFormatter: (v: number) => (v > 0 ? v.toFixed(0) : '–'),
    },
    {
      field: 'cold',
      headerName: `${t('run.latency.cold')} ${t('run.latency.ttft')}`,
      width: 120,
      sortable: false,
      valueGetter: (_v, row) => row.cold?.ttft_s ?? 0,
      valueFormatter: (v: number) => fmtSeconds(v),
    },
    {
      field: 'tokens_exact',
      headerName: 'usage',
      width: 90,
      renderCell: (p) =>
        p.row.tokens_exact ? (
          <Chip size="small" label="exact" sx={{ bgcolor: `${tk.tierA}22`, color: tk.tierA, height: 20 }} />
        ) : (
          <Chip size="small" label="est." title={t('run.latency.estimated')} sx={{ height: 20 }} />
        ),
    },
    {
      field: 'served_model',
      headerName: t('run.models.served'),
      width: 180,
      renderCell: (p) => (
        <span className="mono" style={{ fontSize: 12 }}>
          {p.row.served_model && p.row.served_model !== p.row.model_id ? p.row.served_model : ''}
        </span>
      ),
    },
  ];
  if (q.isLoading) return <LoadingBlock rows={3} />;
  if (q.error) return <EmptyState title={t('common.empty')} description={String(q.error)} />;
  if (q.isSuccess && stats.length === 0) return <EmptyState title={t('common.empty')} />;
  return (
    <Stack spacing={2}>
      <Typography variant="body2" color="text.secondary">
        {t('run.latency.intro')}
      </Typography>
      {best && (
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: 'repeat(3, 1fr)' }, gap: 2 }}>
          <Section title={t('run.latency.fastestTtft')} subtitle={best.byTtft.model_id}>
            <Gauge
              value={best.byTtft.ttft_s.median * 1000}
              max={Math.max(500, Math.ceil((best.maxTtft * 1000) / 500) * 500)}
              label="TTFT p50"
              unit=" ms"
              color={tk.tierA}
            />
          </Section>
          <Section title={t('run.latency.highestTps')} subtitle={best.byTps.model_id}>
            <Gauge
              value={best.byTps.tok_per_s.median}
              max={Math.max(50, Math.ceil(best.maxTps / 50) * 50)}
              label="tok/s p50"
              digits={1}
            />
          </Section>
          <Section title={t('run.latency.lowestE2e')} subtitle={best.byE2e.model_id}>
            <Gauge
              value={best.byE2e.e2e_s.median}
              max={Math.max(1, Math.ceil(best.maxE2e))}
              label="E2E p50"
              unit=" s"
              digits={2}
              color={tk.secondary}
            />
          </Section>
        </Box>
      )}
      <Section title={t('run.tabs.latency')} subtitle={`${stats.length} ${t('run.latency.measured')}`}>
        <LatencyBars stats={stats} />
      </Section>
      <Box
        sx={{
          bgcolor: 'background.paper',
          borderRadius: radius.card,
          border: '1px solid',
          borderColor: 'divider',
          overflow: 'hidden',
        }}
      >
        <DataGrid
          {...gridDefaults}
          rows={q.data?.stats ?? []}
          getRowId={(r) => r.model_id}
          columns={columns}
          loading={q.isLoading}
          localeText={gridLocale(i18n.language)}
          onRowClick={(p) => onModel(String(p.id))}
          autoHeight
          sx={gridSx}
        />
      </Box>
    </Stack>
  );
}
