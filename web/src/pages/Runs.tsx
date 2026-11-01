import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Button,
  Checkbox,
  IconButton,
  Stack,
  ToggleButton,
  ToggleButtonGroup,
  Tooltip,
  Typography,
} from '@mui/material';
import { DataGrid, type GridColDef } from '@mui/x-data-grid';
import DeleteOutlineRounded from '@mui/icons-material/DeleteOutlineRounded';
import StopCircleOutlined from '@mui/icons-material/StopCircleOutlined';
import CompareArrowsRounded from '@mui/icons-material/CompareArrowsRounded';
import BoltRounded from '@mui/icons-material/BoltRounded';
import { useCancelRun, useDeleteRun, useRuns } from '@/api/hooks';
import type { Run } from '@/api/types';
import { PageHeader } from '@/components/PageHeader';
import { EmptyState } from '@/components/EmptyState';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { KindChip, StatusChip } from '@/components/badges';
import { durationBetween, fmtDate } from '@/lib/format';
import { gridDefaults, gridLocale, gridSx } from '@/lib/grid';
import { isLive, runTitle, scanCounts } from '@/lib/run';

const KINDS = ['suite', 'scan', 'latency', 'load', 'capability', 'quality'] as const;

export default function Runs() {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [kind, setKind] = useState<string>('');
  const runs = useRuns(kind || undefined, 200);
  const del = useDeleteRun();
  const cancel = useCancelRun();
  const [toDelete, setToDelete] = useState<Run | null>(null);
  const [selected, setSelected] = useState<string[]>([]);
  const rows = useMemo(() => (runs.data?.runs ?? []).filter((r) => !r.parent_id), [runs.data]);
  const toggle = (id: string) =>
    setSelected((s) => (s.includes(id) ? s.filter((x) => x !== id) : [...s.slice(-1), id]));

  const columns: GridColDef<Run>[] = [
    {
      field: 'select',
      headerName: '',
      width: 48,
      sortable: false,
      renderCell: (p) =>
        (p.row.kind === 'scan' || p.row.kind === 'suite') && p.row.status === 'done' ? (
          <Checkbox
            size="small"
            checked={selected.includes(p.row.id)}
            onClick={(e) => e.stopPropagation()}
            onChange={() => toggle(p.row.id)}
            inputProps={{ 'aria-label': t('compare.title') }}
          />
        ) : null,
    },
    {
      field: 'status',
      headerName: t('common.status'),
      width: 120,
      renderCell: (p) => <StatusChip status={p.row.status} />,
    },
    { field: 'kind', headerName: t('common.kind'), width: 140, renderCell: (p) => <KindChip kind={p.row.kind} /> },
    {
      field: 'endpoint_name',
      headerName: t('common.endpoint'),
      flex: 1,
      minWidth: 200,
      valueGetter: (_v, row) => runTitle(row),
      renderCell: (p) => (
        <Typography variant="body2" sx={{ fontWeight: 600 }} noWrap>
          {runTitle(p.row)}
        </Typography>
      ),
    },
    { field: 'preset', headerName: t('newRun.profile'), width: 110 },
    {
      field: 'counts',
      headerName: `${t('runs.available')} / ${t('common.models')}`,
      width: 150,
      sortable: false,
      renderCell: (p) => {
        const c = scanCounts(p.row);
        return c ? (
          <span className="mono">
            {c.ok} / {c.models}
          </span>
        ) : (
          <span>–</span>
        );
      },
    },
    { field: 'created_at', headerName: t('common.createdAt'), width: 170, valueFormatter: (v: string) => fmtDate(v) },
    {
      field: 'duration',
      headerName: t('common.duration'),
      width: 100,
      sortable: false,
      valueGetter: (_v, row) => durationBetween(row.started_at, row.finished_at),
    },
    { field: 'dialect', headerName: 'Dialect', width: 110 },
    {
      field: 'actions',
      headerName: t('common.actions'),
      width: 100,
      sortable: false,
      renderCell: (p) => (
        <Stack direction="row" onClick={(e) => e.stopPropagation()}>
          {isLive(p.row.status) ? (
            <Tooltip title={t('runs.cancel')}>
              <IconButton size="small" onClick={() => cancel.mutate(p.row.id)}>
                <StopCircleOutlined fontSize="small" />
              </IconButton>
            </Tooltip>
          ) : (
            <Tooltip title={t('common.delete')}>
              <IconButton size="small" onClick={() => setToDelete(p.row)}>
                <DeleteOutlineRounded fontSize="small" />
              </IconButton>
            </Tooltip>
          )}
        </Stack>
      ),
    },
  ];

  return (
    <>
      <PageHeader
        title={t('runs.title')}
        subtitle={t('runs.subtitle')}
        actions={
          <>
            <Button
              variant="outlined"
              startIcon={<CompareArrowsRounded />}
              disabled={selected.length !== 2}
              onClick={() => void navigate(`/compare?a=${selected[0]}&b=${selected[1]}`)}
            >
              {t('runs.compareSelected')}
            </Button>
            <Button
              variant="contained"
              disableElevation
              startIcon={<BoltRounded />}
              onClick={() => void navigate('/new')}
            >
              {t('nav.newRun')}
            </Button>
          </>
        }
      />
      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 2 }} flexWrap="wrap" useFlexGap>
        <ToggleButtonGroup size="small" exclusive value={kind} onChange={(_e, v: string | null) => setKind(v ?? '')}>
          <ToggleButton value="">{t('common.all')}</ToggleButton>
          {KINDS.map((k) => (
            <ToggleButton key={k} value={k}>
              {t(`kinds.${k}`)}
            </ToggleButton>
          ))}
        </ToggleButtonGroup>
        <Typography variant="body2" color="text.secondary">
          {t('runs.selectTwo')}
        </Typography>
      </Stack>
      {runs.isSuccess && rows.length === 0 ? (
        <EmptyState
          title={t('runs.none')}
          action={
            <Button variant="contained" disableElevation onClick={() => void navigate('/new')}>
              {t('dashboard.start')}
            </Button>
          }
        />
      ) : (
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
            rows={rows}
            columns={columns}
            loading={runs.isLoading}
            localeText={gridLocale(i18n.language)}
            onRowClick={(p) => void navigate(`/runs/${p.id}`)}
            autoHeight
            sx={gridSx}
          />
        </Box>
      )}
      <ConfirmDialog
        open={!!toDelete}
        danger
        title={t('common.delete')}
        text={toDelete ? t('runs.deleteConfirm', { id: toDelete.id }) : ''}
        confirmLabel={t('common.delete')}
        busy={del.isPending}
        onClose={() => setToDelete(null)}
        onConfirm={() => toDelete && del.mutate(toDelete.id, { onSuccess: () => setToDelete(null) })}
      />
    </>
  );
}
