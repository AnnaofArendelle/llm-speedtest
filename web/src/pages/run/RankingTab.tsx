import { useMemo } from 'react';
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Box,
  Chip,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material';
import ExpandMoreRounded from '@mui/icons-material/ExpandMoreRounded';
import { DataGrid, type GridColDef } from '@mui/x-data-grid';
import { useTranslation } from 'react-i18next';
import { useFamilies, useSummary } from '@/api/hooks';
import type { FamilyStat } from '@/api/types';
import { Section } from '@/components/Section';
import { EmptyState } from '@/components/EmptyState';
import { LatencyChip, ReasonChip, ScoreBar, TierBadge } from '@/components/badges';
import { useTokens } from '@/lib/tokens';
import { tierColor } from '@/app/theme';
import { fmtSeconds } from '@/lib/format';
import { gridDefaults, gridLocale, gridSx } from '@/lib/grid';
import type { TabProps } from './types';

export function RankingTab({ runId, onFamily, onModel, order }: TabProps & { order: 'composite' | 'stability' }) {
  const { t, i18n } = useTranslation();
  const tk = useTokens();
  const q = useFamilies(runId, order);
  const sum = useSummary(runId);
  const scoring = sum.data?.scoring;
  const rows = useMemo(() => q.data?.families ?? [], [q.data]);
  const unavailable = q.data?.unavailable ?? [];
  const rankField = order === 'composite' ? 'rank_composite' : 'rank_stability';

  const columns: GridColDef<FamilyStat>[] = [
    {
      field: rankField,
      headerName: '#',
      width: 60,
      valueGetter: (_v, row) => (order === 'composite' ? row.rank_composite : row.rank_stability) ?? 0,
      renderCell: (p) => <span className="mono">{p.value as number}</span>,
    },
    {
      field: 'display',
      headerName: t('common.family'),
      flex: 1,
      minWidth: 200,
      renderCell: (p) => (
        <Typography variant="body2" sx={{ fontWeight: 700 }} noWrap title={p.row.key}>
          {p.row.display}
        </Typography>
      ),
    },
    { field: 'tier', headerName: t('common.stability'), width: 90, renderCell: (p) => <TierBadge tier={p.row.tier} /> },
    ...(order === 'composite'
      ? ([
          {
            field: 'score',
            headerName: t('common.score'),
            width: 170,
            renderCell: (p) => <ScoreBar value={p.row.score} color={tierColor(tk, p.row.tier)} />,
          },
          {
            field: 's_stability',
            headerName: `${t('run.ranking.stabilityScore')} · ${t('run.ranking.breadth')} · ${t('run.ranking.speed')}`,
            width: 170,
            sortable: false,
            renderCell: (p) => (
              <span className="mono" style={{ fontSize: 12 }}>
                {p.row.s_stability.toFixed(0)} · {p.row.s_breadth.toFixed(0)} · {p.row.s_speed.toFixed(0)}
              </span>
            ),
          },
        ] as GridColDef<FamilyStat>[])
      : ([
          {
            field: 'rate',
            headerName: t('common.stability'),
            width: 110,
            valueFormatter: (v: number) => `${(v * 100).toFixed(1)}%`,
          },
          {
            field: 'attempts',
            headerName: t('common.attempts'),
            width: 110,
            renderCell: (p) => (
              <span className="mono">
                {p.row.successes}/{p.row.attempts}
              </span>
            ),
          },
        ] as GridColDef<FamilyStat>[])),
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
      width: 220,
      renderCell: (p) => (
        <Stack direction="row" spacing={0.5} sx={{ overflow: 'hidden' }}>
          <span className="mono">{p.row.n_prov_ok}</span>
          {Object.entries(p.row.ok_providers)
            .sort((a, b) => b[1] - a[1])
            .slice(0, 3)
            .map(([name]) => (
              <Chip key={name} size="small" label={name} variant="outlined" sx={{ height: 20, fontSize: 11 }} />
            ))}
          {Object.keys(p.row.ok_providers).length > 3 && (
            <Chip
              size="small"
              label={`+${Object.keys(p.row.ok_providers).length - 3}`}
              sx={{ height: 20, fontSize: 11 }}
            />
          )}
        </Stack>
      ),
    },
    {
      field: 'median_s',
      headerName: t('common.median'),
      width: 110,
      renderCell: (p) => <LatencyChip seconds={p.row.median_s} />,
    },
    {
      field: 'min_s',
      headerName: `${t('common.min')} – ${t('common.max')}`,
      width: 150,
      renderCell: (p) => (
        <span className="mono">
          {fmtSeconds(p.row.min_s)} – {fmtSeconds(p.row.max_s)}
        </span>
      ),
    },
    {
      field: 'best_id',
      headerName: t('run.ranking.stableMembers'),
      width: 240,
      renderCell: (p) => (
        <Typography
          variant="body2"
          className="mono"
          noWrap
          title={(p.row.stable_ids ?? []).join('\n')}
          onClick={(e) => {
            e.stopPropagation();
            if (p.row.best_id) onModel(p.row.best_id);
          }}
          sx={{ cursor: 'pointer', textDecoration: 'underline dotted' }}
        >
          {p.row.best_id ?? '–'}
          {(p.row.stable_ids?.length ?? 0) > 1 ? ` +${(p.row.stable_ids?.length ?? 1) - 1}` : ''}
        </Typography>
      ),
    },
    {
      field: 'reasons',
      headerName: t('run.ranking.topReasons'),
      width: 200,
      sortable: false,
      renderCell: (p) => {
        const top = Object.entries(p.row.reasons).sort((a, b) => b[1] - a[1]);
        if (!top.length) return <span>–</span>;
        return (
          <Stack direction="row" spacing={0.5} alignItems="center">
            <ReasonChip code={top[0][0]} />
            <span className="mono" style={{ fontSize: 12 }}>
              ×{top[0][1]}
            </span>
            {top.length > 1 && <span style={{ fontSize: 12, opacity: 0.7 }}>+{top.length - 1}</span>}
          </Stack>
        );
      },
    },
  ];

  if (q.error) return <EmptyState title={t('run.notScan')} description={String(q.error)} />;
  return (
    <Stack spacing={2}>
      {scoring && order === 'composite' && (
        <Typography variant="body2" color="text.secondary">
          {t('run.ranking.intro', { ws: scoring.w_stability, wb: scoring.w_breadth, wsp: scoring.w_speed })}
        </Typography>
      )}
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
      {order === 'composite' && unavailable.length > 0 && (
        <Accordion disableGutters sx={{ borderRadius: 5, '&:before': { display: 'none' }, overflow: 'hidden' }}>
          <AccordionSummary expandIcon={<ExpandMoreRounded />}>
            <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
              {t('run.ranking.unavailable')} ({unavailable.length})
            </Typography>
          </AccordionSummary>
          <AccordionDetails sx={{ p: 0 }}>
            <Section noPad>
              <Box sx={{ overflowX: 'auto' }}>
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>{t('common.family')}</TableCell>
                      <TableCell>{t('run.ranking.members')}</TableCell>
                      <TableCell>{t('common.providers')}</TableCell>
                      <TableCell>{t('run.ranking.topReasons')}</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {unavailable.map((f) => (
                      <TableRow key={f.key} hover sx={{ cursor: 'pointer' }} onClick={() => onFamily(f.key)}>
                        <TableCell sx={{ fontWeight: 600 }}>{f.display}</TableCell>
                        <TableCell className="mono">{f.n}</TableCell>
                        <TableCell>
                          {Object.entries(f.providers)
                            .map(([p, n]) => `${p} (${n})`)
                            .join(', ')}
                        </TableCell>
                        <TableCell>
                          <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap>
                            {Object.entries(f.reasons)
                              .sort((a, b) => b[1] - a[1])
                              .slice(0, 3)
                              .map(([r, n]) => (
                                <Stack key={r} direction="row" spacing={0.25} alignItems="center">
                                  <ReasonChip code={r} />
                                  <span className="mono" style={{ fontSize: 12 }}>
                                    ×{n}
                                  </span>
                                </Stack>
                              ))}
                          </Stack>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </Box>
            </Section>
          </AccordionDetails>
        </Accordion>
      )}
    </Stack>
  );
}
