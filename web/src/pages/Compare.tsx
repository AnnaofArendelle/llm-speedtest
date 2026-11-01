import { useMemo } from 'react';
import { useSearchParams } from 'react-router';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Chip,
  FormControl,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  Tooltip,
  Typography,
} from '@mui/material';
import { DataGrid, type GridColDef } from '@mui/x-data-grid';
import SwapHorizRounded from '@mui/icons-material/SwapHorizRounded';
import { useCompare, useRuns } from '@/api/hooks';
import type { CompareDoc } from '@/api/types';
import { PageHeader } from '@/components/PageHeader';
import { EmptyState } from '@/components/EmptyState';
import { Section } from '@/components/Section';
import { KpiCard } from '@/components/KpiCard';
import { LatencyChip, TierBadge } from '@/components/badges';
import { useTokens } from '@/lib/tokens';
import { fmtDate, fmtPct, fmtSeconds } from '@/lib/format';
import { gridDefaults, gridLocale, gridSx } from '@/lib/grid';
import { runTitle } from '@/lib/run';

type ModelDiff = CompareDoc['models'][number];
type FamilyDiff = CompareDoc['families'][number];

export default function Compare() {
  const { t, i18n } = useTranslation();
  const tk = useTokens();
  const [params, setParams] = useSearchParams();
  const a = params.get('a') ?? '';
  const b = params.get('b') ?? '';
  const runs = useRuns(undefined, 200);
  const candidates = useMemo(
    () =>
      (runs.data?.runs ?? []).filter(
        (r) => !r.parent_id && (r.kind === 'scan' || r.kind === 'suite') && r.status === 'done',
      ),
    [runs.data],
  );
  const cmp = useCompare(a || undefined, b || undefined);
  const set = (k: 'a' | 'b', v: string) =>
    setParams(
      (prev) => {
        prev.set(k, v);
        return prev;
      },
      { replace: true },
    );
  const swap = () =>
    setParams(
      (prev) => {
        prev.set('a', b);
        prev.set('b', a);
        return prev;
      },
      { replace: true },
    );
  const changeColor: Record<string, string> = {
    improved: tk.tierA,
    regressed: tk.fail,
    new: tk.secondary,
    gone: tk.tierC,
    same: tk.variant,
  };
  const d = cmp.data;

  const modelCols: GridColDef<ModelDiff>[] = [
    {
      field: 'change',
      headerName: t('common.status'),
      width: 110,
      renderCell: (p) => (
        <Chip
          size="small"
          label={t(`compare.${p.row.change}`, { defaultValue: p.row.change })}
          sx={{ bgcolor: `${changeColor[p.row.change]}22`, color: changeColor[p.row.change], fontWeight: 700 }}
        />
      ),
    },
    {
      field: 'id',
      headerName: t('common.model'),
      flex: 1,
      minWidth: 220,
      renderCell: (p) => (
        <span className="mono" title={p.row.id}>
          {p.row.id}
        </span>
      ),
    },
    { field: 'provider', headerName: t('common.provider'), width: 130 },
    { field: 'family', headerName: t('common.family'), width: 150 },
    {
      field: 'status',
      headerName: `A → B`,
      width: 130,
      sortable: false,
      renderCell: (p) => (
        <Stack direction="row" spacing={0.5} alignItems="center">
          {p.row.status_a === 'ok' ? (
            <TierBadge tier={p.row.tier_a} />
          ) : (
            <Chip size="small" label={p.row.status_a || '–'} />
          )}
          <span>→</span>
          {p.row.status_b === 'ok' ? (
            <TierBadge tier={p.row.tier_b} />
          ) : (
            <Chip size="small" label={p.row.status_b || '–'} />
          )}
        </Stack>
      ),
    },
    {
      field: 'median_a',
      headerName: `${t('common.median')} A`,
      width: 110,
      renderCell: (p) => (p.row.median_a > 0 ? <LatencyChip seconds={p.row.median_a} /> : '–'),
    },
    {
      field: 'median_b',
      headerName: `${t('common.median')} B`,
      width: 110,
      renderCell: (p) => (p.row.median_b > 0 ? <LatencyChip seconds={p.row.median_b} /> : '–'),
    },
    {
      field: 'delta_median',
      headerName: t('compare.deltaMedian'),
      width: 120,
      renderCell: (p) =>
        p.row.median_a > 0 && p.row.median_b > 0 ? (
          <span className="mono" style={{ color: p.row.delta_median > 0 ? tk.fail : tk.tierA }}>
            {p.row.delta_median > 0 ? '+' : ''}
            {p.row.delta_median.toFixed(2)}s
          </span>
        ) : (
          '–'
        ),
    },
    {
      field: 'reason_b',
      headerName: `${t('common.reason')} B`,
      width: 160,
      renderCell: (p) => (
        <span className="mono">{p.row.reason_b && p.row.reason_b !== 'ok' ? p.row.reason_b : '–'}</span>
      ),
    },
  ];
  const familyCols: GridColDef<FamilyDiff>[] = [
    { field: 'display', headerName: t('common.family'), flex: 1, minWidth: 180 },
    {
      field: 'score_a',
      headerName: `${t('compare.score')} A`,
      width: 110,
      valueFormatter: (v: number) => v.toFixed(1),
    },
    {
      field: 'score_b',
      headerName: `${t('compare.score')} B`,
      width: 110,
      valueFormatter: (v: number) => v.toFixed(1),
    },
    {
      field: 'delta',
      headerName: 'Δ',
      width: 90,
      valueGetter: (_v, row) => row.score_b - row.score_a,
      renderCell: (p) => (
        <span className="mono" style={{ color: p.row.score_b - p.row.score_a >= 0 ? tk.tierA : tk.fail }}>
          {(p.row.score_b - p.row.score_a).toFixed(1)}
        </span>
      ),
    },
    {
      field: 'rank_a',
      headerName: `${t('common.rank')} A`,
      width: 90,
      valueFormatter: (v: number) => (v > 0 ? `#${v}` : '–'),
    },
    {
      field: 'rank_b',
      headerName: `${t('common.rank')} B`,
      width: 90,
      valueFormatter: (v: number) => (v > 0 ? `#${v}` : '–'),
    },
    { field: 'ok_a', headerName: `${t('runs.available')} A`, width: 100 },
    { field: 'ok_b', headerName: `${t('runs.available')} B`, width: 100 },
    {
      field: 'median_a',
      headerName: `${t('common.median')} A`,
      width: 110,
      valueFormatter: (v: number) => fmtSeconds(v),
    },
    {
      field: 'median_b',
      headerName: `${t('common.median')} B`,
      width: 110,
      valueFormatter: (v: number) => fmtSeconds(v),
    },
  ];

  const picker = (k: 'a' | 'b', value: string) => (
    <FormControl size="small" sx={{ minWidth: 260, flex: 1 }}>
      <InputLabel>{t(`compare.${k}`)}</InputLabel>
      <Select
        label={t(`compare.${k}`)}
        value={candidates.some((r) => r.id === value) ? value : ''}
        onChange={(e) => set(k, e.target.value)}
      >
        {candidates.map((r) => (
          <MenuItem key={r.id} value={r.id}>
            {runTitle(r)} · {r.preset} · {fmtDate(r.finished_at || r.created_at)}
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );

  return (
    <>
      <PageHeader title={t('compare.title')} subtitle={t('compare.subtitle')} />
      <Stack direction={{ xs: 'column', md: 'row' }} spacing={1} alignItems="center" sx={{ mb: 2 }}>
        {picker('a', a)}
        <Tooltip title={t('compare.swap')}>
          <IconButton onClick={swap} disabled={!a || !b}>
            <SwapHorizRounded />
          </IconButton>
        </Tooltip>
        {picker('b', b)}
      </Stack>
      {!a || !b ? (
        <EmptyState title={t('compare.pick')} />
      ) : cmp.error ? (
        <EmptyState title={t('common.error')} description={String(cmp.error)} />
      ) : d ? (
        <Stack spacing={2}>
          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr 1fr', md: 'repeat(5, 1fr)' }, gap: 2 }}>
            {(['improved', 'regressed', 'new', 'gone', 'same'] as const).map((k) => (
              <KpiCard key={k} label={t(`compare.${k}`)} value={d.counts[k] ?? 0} color={changeColor[k]} />
            ))}
          </Box>
          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2 }}>
            {(['a', 'b'] as const).map((k) => {
              const run = k === 'a' ? d.a : d.b;
              const s = k === 'a' ? d.summary_a : d.summary_b;
              return (
                <Section
                  key={k}
                  title={`${t(`compare.${k}`)} · ${runTitle(run)}`}
                  subtitle={`${run.preset} · ${fmtDate(run.finished_at || run.created_at)}`}
                >
                  <Stack direction="row" spacing={3} flexWrap="wrap" useFlexGap>
                    <Typography variant="body2">
                      {t('common.models')}: <b className="mono">{s.models}</b>
                    </Typography>
                    <Typography variant="body2">
                      {t('runs.available')}: <b className="mono">{s.ok}</b> ({fmtPct(s.ok, s.models)})
                    </Typography>
                    <Typography variant="body2">
                      {t('common.median')}: <b className="mono">{fmtSeconds(s.latency_median_s)}</b>
                    </Typography>
                    <Typography variant="body2">
                      {t('common.families')}:{' '}
                      <b className="mono">
                        {s.families_ok}/{s.families}
                      </b>
                    </Typography>
                  </Stack>
                </Section>
              );
            })}
          </Box>
          <Section title={t('compare.models')} noPad>
            <DataGrid
              {...gridDefaults}
              rows={d.models}
              columns={modelCols}
              localeText={gridLocale(i18n.language)}
              autoHeight
              sx={gridSx}
            />
          </Section>
          <Section title={t('compare.familyDiff')} noPad>
            <DataGrid
              {...gridDefaults}
              rows={d.families}
              getRowId={(r) => r.key}
              columns={familyCols}
              localeText={gridLocale(i18n.language)}
              autoHeight
              sx={gridSx}
              initialState={{
                pagination: { paginationModel: { pageSize: 25 } },
                sorting: { sortModel: [{ field: 'delta', sort: 'asc' }] },
              }}
            />
          </Section>
        </Stack>
      ) : null}
    </>
  );
}
