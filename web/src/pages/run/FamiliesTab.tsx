import { useMemo, useState } from 'react';
import { Box, Stack, TextField, Typography } from '@mui/material';
import { DataGrid, type GridColDef } from '@mui/x-data-grid';
import { useTranslation } from 'react-i18next';
import { radius } from '@/app/theme';
import { useFamilies } from '@/api/hooks';
import type { FamilyStat } from '@/api/types';
import { EmptyState } from '@/components/EmptyState';
import { LatencyChip, TierBadge } from '@/components/badges';
import { gridDefaults, gridLocale, gridSx } from '@/lib/grid';
import { arr, entries, keys } from '@/lib/safe';
import type { TabProps } from './types';

export default function FamiliesTab({ runId, onFamily }: TabProps) {
  const { t, i18n } = useTranslation();
  const q = useFamilies(runId, 'composite');
  const [search, setSearch] = useState('');
  const rows = useMemo(() => {
    const all = [...(q.data?.families ?? []), ...(q.data?.unavailable ?? [])];
    const s = search.trim().toLowerCase();
    return s
      ? all.filter(
          (f) =>
            f.display.toLowerCase().includes(s) ||
            f.key.includes(s) ||
            arr(f.member_ids).some((m) => m.toLowerCase().includes(s)),
        )
      : all;
  }, [q.data, search]);
  const columns: GridColDef<FamilyStat>[] = [
    {
      field: 'display',
      headerName: t('common.family'),
      flex: 1,
      minWidth: 200,
      renderCell: (p) => (
        <Typography variant="body2" sx={{ fontWeight: 700 }} noWrap>
          {p.row.display}
        </Typography>
      ),
    },
    {
      field: 'key',
      headerName: t('run.families.key'),
      width: 220,
      renderCell: (p) => (
        <span className="mono" style={{ fontSize: 12 }}>
          {p.row.key}
        </span>
      ),
    },
    { field: 'tier', headerName: t('common.stability'), width: 90, renderCell: (p) => <TierBadge tier={p.row.tier} /> },
    {
      field: 'ok',
      headerName: t('run.ranking.availableMembers'),
      width: 110,
      renderCell: (p) => (
        <span className="mono">
          {p.row.ok}/{p.row.n}
        </span>
      ),
    },
    {
      field: 'n_prov_ok',
      headerName: t('run.ranking.providersOk'),
      width: 110,
      renderCell: (p) => (
        <span className="mono">
          {p.row.n_prov_ok}/{keys(p.row.providers).length}
        </span>
      ),
    },
    {
      field: 'median_s',
      headerName: t('common.median'),
      width: 110,
      renderCell: (p) => (p.row.median_s > 0 ? <LatencyChip seconds={p.row.median_s} /> : '–'),
    },
    {
      field: 'score',
      headerName: t('common.score'),
      width: 90,
      valueFormatter: (v: number) => (v > 0 ? v.toFixed(1) : '–'),
    },
    {
      field: 'sources',
      headerName: t('run.families.sources'),
      width: 220,
      sortable: false,
      renderCell: (p) => (
        <span className="mono" style={{ fontSize: 12 }}>
          {entries(p.row.sources)
            .map(([s, n]) => `${s}×${n}`)
            .join(' ')}
        </span>
      ),
    },
  ];
  if (q.error) return <EmptyState title={t('run.notScan')} description={String(q.error)} />;
  return (
    <Stack spacing={2}>
      <Stack direction="row" spacing={1} alignItems="center">
        <TextField
          size="small"
          label={t('run.families.search')}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          sx={{ minWidth: 260 }}
        />
        <Typography variant="body2" color="text.secondary">
          {rows.length} {t('common.families')}
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
          initialState={{ pagination: { paginationModel: { pageSize: 50 } } }}
          rows={rows}
          getRowId={(r) => r.key}
          columns={columns}
          loading={q.isLoading}
          localeText={gridLocale(i18n.language)}
          onRowClick={(p) => onFamily(String(p.id))}
          autoHeight
          sx={gridSx}
        />
      </Box>
    </Stack>
  );
}
