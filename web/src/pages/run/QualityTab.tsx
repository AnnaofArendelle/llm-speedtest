import { Box, Chip, Stack, Typography } from '@mui/material';
import { DataGrid, type GridColDef } from '@mui/x-data-grid';
import { useTranslation } from 'react-i18next';
import { radius } from '@/app/theme';
import { useQuality } from '@/api/hooks';
import type { QualityScore } from '@/api/types';
import { EmptyState } from '@/components/EmptyState';
import { ScoreBar } from '@/components/badges';
import { useTokens } from '@/lib/tokens';
import { truncate } from '@/lib/format';
import { gridDefaults, gridLocale, gridSx } from '@/lib/grid';
import type { TabProps } from './types';

export default function QualityTab({ runId, onModel }: TabProps) {
  const { t, i18n } = useTranslation();
  const tk = useTokens();
  const q = useQuality(runId);
  const scores = q.data?.scores ?? [];
  const categories = Array.from(new Set(scores.flatMap((s) => Object.keys(s.categories ?? {})))).sort();
  const columns: GridColDef<QualityScore>[] = [
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
      field: 'score',
      headerName: t('run.quality.score'),
      width: 170,
      renderCell: (p) => (
        <ScoreBar value={p.row.score} color={p.row.score >= 80 ? tk.tierA : p.row.score >= 50 ? tk.tierB : tk.fail} />
      ),
    },
    {
      field: 'consistency',
      headerName: t('run.quality.consistency'),
      width: 110,
      valueFormatter: (v: number) => `${(v * 100).toFixed(0)}%`,
    },
    {
      field: 'answered',
      headerName: t('run.quality.answered'),
      width: 90,
      renderCell: (p) => (
        <span className="mono">
          {p.row.answered}/{Object.keys(p.row.questions ?? {}).length * p.row.trials}
        </span>
      ),
    },
    {
      field: 'trials',
      headerName: t('run.quality.trials'),
      width: 90,
      renderCell: (p) => (
        <span className="mono">{(p.row.trial_scores ?? []).map((x) => x.toFixed(0)).join(' / ')}</span>
      ),
    },
    ...categories.map((c): GridColDef<QualityScore> => ({
      field: `cat_${c}`,
      headerName: c,
      width: 90,
      valueGetter: (_v, row) => row.categories?.[c] ?? 0,
      renderCell: (p) => {
        const v = p.row.categories?.[c];
        if (v === undefined) return '–';
        const color = v >= 0.8 ? tk.tierA : v >= 0.5 ? tk.tierB : tk.fail;
        return (
          <Chip
            size="small"
            label={`${(v * 100).toFixed(0)}%`}
            sx={{ bgcolor: `${color}22`, color, height: 20, fontWeight: 700 }}
          />
        );
      },
    })),
    {
      field: 'identity',
      headerName: t('run.quality.identity'),
      width: 260,
      renderCell: (p) => (
        <span title={p.row.identity} style={{ fontSize: 12 }}>
          {truncate(p.row.identity, 60)}
        </span>
      ),
    },
    {
      field: 'errors',
      headerName: t('common.reason'),
      width: 160,
      sortable: false,
      renderCell: (p) => (
        <span className="mono" style={{ fontSize: 12 }}>
          {Object.entries(p.row.errors ?? {})
            .map(([k, v]) => `${k}×${v}`)
            .join(' ')}
        </span>
      ),
    },
  ];
  if (q.error) return <EmptyState title={t('common.empty')} description={String(q.error)} />;
  if (q.isSuccess && scores.length === 0) return <EmptyState title={t('common.empty')} />;
  return (
    <Stack spacing={2}>
      <Typography variant="body2" color="text.secondary">
        {t('run.quality.intro')}
      </Typography>
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
          rows={scores}
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
