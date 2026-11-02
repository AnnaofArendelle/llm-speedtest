import { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router';
import { Box, FormControl, InputLabel, MenuItem, Select, Stack, TextField, Typography } from '@mui/material';
import { DataGrid, type GridColDef, type GridPaginationModel, type GridSortModel } from '@mui/x-data-grid';
import { useTranslation } from 'react-i18next';
import { radius } from '@/app/theme';
import { useModels, useProviders, useSummary } from '@/api/hooks';
import type { ModelStat } from '@/api/types';
import { EmptyState } from '@/components/EmptyState';
import { LatencyChip, ProviderChip, ReasonChip, TierBadge } from '@/components/badges';
import { fmtContext, fmtSeconds } from '@/lib/format';
import { gridDefaults, gridLocale, gridSx } from '@/lib/grid';
import { useReasonLookup } from '@/lib/reasons';
import { keys, obj } from '@/lib/safe';
import type { TabProps } from './types';

const SORT_KEYS: Record<string, string> = {
  id: 'id',
  provider_name: 'provider',
  family_canon: 'family',
  median_s: 'median',
  min_s: 'min',
  max_s: 'max',
  rate: 'rate',
  attempts: 'attempts',
  tier: 'status',
  reason: 'reason',
};

export function ModelsTab({ runId, onModel }: TabProps) {
  const { t, i18n } = useTranslation();
  const lookup = useReasonLookup();
  const [params, setParams] = useSearchParams();
  const status = params.get('status') ?? '';
  const tier = params.get('tier') ?? '';
  const provider = params.get('provider') ?? '';
  const reason = params.get('reason') ?? '';
  const [search, setSearch] = useState(params.get('q') ?? '');
  const [debounced, setDebounced] = useState(search);
  const [pm, setPm] = useState<GridPaginationModel>({ page: 0, pageSize: 50 });
  const [sort, setSort] = useState<GridSortModel>([]);
  useEffect(() => {
    const h = window.setTimeout(() => setDebounced(search), 250);
    return () => window.clearTimeout(h);
  }, [search]);
  const setFilter = (k: string, v: string) => {
    setParams(
      (prev) => {
        if (v) prev.set(k, v);
        else prev.delete(k);
        return prev;
      },
      { replace: true },
    );
    setPm((p) => ({ ...p, page: 0 }));
  };
  const sortKey = sort[0] ? (SORT_KEYS[sort[0].field] ?? '') : '';
  const query = useMemo(() => {
    const u = new URLSearchParams();
    if (status) u.set('status', status);
    if (tier) u.set('tier', tier);
    if (provider) u.set('provider', provider);
    if (reason) u.set('reason', reason);
    if (debounced) u.set('q', debounced);
    if (sortKey) {
      u.set('sort', sortKey);
      u.set('order', sort[0]?.sort ?? 'asc');
    }
    u.set('page', String(pm.page + 1));
    u.set('size', String(pm.pageSize));
    return u.toString();
  }, [status, tier, provider, reason, debounced, sortKey, sort, pm]);
  const q = useModels(runId, query);
  const providers = useProviders(runId);
  const summary = useSummary(runId);
  const reasons = keys(summary.data?.summary?.reasons).sort();

  const columns: GridColDef<ModelStat>[] = [
    {
      field: 'id',
      headerName: t('common.model'),
      flex: 1,
      minWidth: 260,
      renderCell: (p) => (
        <Typography variant="body2" className="mono" noWrap title={p.row.id}>
          {p.row.id}
        </Typography>
      ),
    },
    {
      field: 'provider_name',
      headerName: t('common.provider'),
      width: 140,
      renderCell: (p) => <ProviderChip name={p.row.provider_name} />,
    },
    {
      field: 'family_canon',
      headerName: t('common.family'),
      width: 170,
      renderCell: (p) => (
        <Typography variant="body2" noWrap title={p.row.family_key}>
          {p.row.family_canon}
        </Typography>
      ),
    },
    {
      field: 'tier',
      headerName: t('common.status'),
      width: 190,
      renderCell: (p) =>
        p.row.status === 'ok' ? (
          <Stack direction="row" spacing={0.5} alignItems="center">
            <TierBadge tier={p.row.tier} />
            <span className="mono" style={{ fontSize: 12 }}>
              {p.row.ok}/{p.row.attempts}
            </span>
          </Stack>
        ) : (
          <ReasonChip code={p.row.reason} />
        ),
    },
    {
      field: 'median_s',
      headerName: t('common.median'),
      width: 110,
      renderCell: (p) =>
        p.row.status === 'ok' ? (
          <LatencyChip seconds={p.row.median_s} />
        ) : (
          <span className="mono" style={{ opacity: 0.6 }}>
            {p.row.fail_latency_s ? fmtSeconds(p.row.fail_latency_s) : '–'}
          </span>
        ),
    },
    { field: 'min_s', headerName: t('common.min'), width: 90, valueFormatter: (v: number) => fmtSeconds(v) },
    { field: 'max_s', headerName: t('common.max'), width: 90, valueFormatter: (v: number) => fmtSeconds(v) },
    {
      field: 'jitter',
      headerName: t('run.models.jitter'),
      width: 90,
      sortable: false,
      valueFormatter: (v: number) => (v > 0 ? `${v.toFixed(2)}×` : '–'),
    },
    {
      field: 'rate',
      headerName: t('common.stability'),
      width: 100,
      valueFormatter: (v: number) => `${(v * 100).toFixed(0)}%`,
    },
    {
      field: 'served_model',
      headerName: t('run.models.served'),
      width: 200,
      sortable: false,
      renderCell: (p) => (
        <Typography variant="body2" className="mono" noWrap title={p.row.served_model}>
          {p.row.served_model && p.row.served_model !== p.row.id && p.row.served_model !== p.row.upstream
            ? p.row.served_model
            : ''}
        </Typography>
      ),
    },
    {
      field: 'context_length',
      headerName: t('run.models.context'),
      width: 90,
      sortable: false,
      valueFormatter: (v: number | undefined) => fmtContext(v),
    },
    {
      field: 'http',
      headerName: 'HTTP',
      width: 90,
      sortable: false,
      renderCell: (p) => (
        <span className="mono">{p.row.http ? `${p.row.http}${p.row.inner ? `/${p.row.inner}` : ''}` : ''}</span>
      ),
    },
  ];
  if (q.error) return <EmptyState title={t('run.notScan')} description={String(q.error)} />;
  return (
    <Stack spacing={2}>
      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
        <TextField
          size="small"
          label={t('common.search')}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          sx={{ minWidth: 220 }}
        />
        <FormControl size="small" sx={{ minWidth: 130 }}>
          <InputLabel>{t('run.models.filterStatus')}</InputLabel>
          <Select
            label={t('run.models.filterStatus')}
            value={status}
            onChange={(e) => setFilter('status', e.target.value)}
          >
            <MenuItem value="">{t('common.all')}</MenuItem>
            <MenuItem value="ok">{t('runs.available')}</MenuItem>
            <MenuItem value="fail">{t('tier.none')}</MenuItem>
          </Select>
        </FormControl>
        <FormControl size="small" sx={{ minWidth: 110 }}>
          <InputLabel>{t('run.models.filterTier')}</InputLabel>
          <Select label={t('run.models.filterTier')} value={tier} onChange={(e) => setFilter('tier', e.target.value)}>
            <MenuItem value="">{t('common.all')}</MenuItem>
            {['A', 'B', 'C'].map((x) => (
              <MenuItem key={x} value={x}>
                {t(`tier.${x}`)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <FormControl size="small" sx={{ minWidth: 180 }}>
          <InputLabel>{t('run.models.filterProvider')}</InputLabel>
          <Select
            label={t('run.models.filterProvider')}
            value={provider}
            onChange={(e) => setFilter('provider', e.target.value)}
          >
            <MenuItem value="">{t('common.all')}</MenuItem>
            {(providers.data?.providers ?? []).map((p) => (
              <MenuItem key={p.key} value={p.key}>
                {p.name} ({p.ok}/{p.models})
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <FormControl size="small" sx={{ minWidth: 200 }}>
          <InputLabel>{t('run.models.filterReason')}</InputLabel>
          <Select
            label={t('run.models.filterReason')}
            value={reason}
            onChange={(e) => setFilter('reason', e.target.value)}
          >
            <MenuItem value="">{t('common.all')}</MenuItem>
            {reasons.map((r) => (
              <MenuItem key={r} value={r}>
                {lookup.name(r)} ({obj(summary.data?.summary?.reasons)[r]})
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <Typography variant="body2" color="text.secondary" sx={{ alignSelf: 'center' }}>
          {q.data ? `${q.data.total} ${t('common.models')}` : ''}
        </Typography>
      </Stack>
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
          pageSizeOptions={[25, 50, 100, 500]}
          rows={q.data?.models ?? []}
          rowCount={q.data?.total ?? 0}
          columns={columns}
          loading={q.isLoading}
          paginationMode="server"
          paginationModel={pm}
          onPaginationModelChange={setPm}
          sortingMode="server"
          sortModel={sort}
          onSortModelChange={setSort}
          localeText={gridLocale(i18n.language)}
          onRowClick={(p) => onModel(String(p.id))}
          autoHeight
          sx={gridSx}
        />
      </Box>
    </Stack>
  );
}
