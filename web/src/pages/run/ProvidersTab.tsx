import { useMemo, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Tooltip,
  Typography,
} from '@mui/material';
import { DataGrid, type GridColDef } from '@mui/x-data-grid';
import { useTranslation } from 'react-i18next';
import { useProviders, useSetProviderOverride } from '@/api/hooks';
import type { ProviderStat } from '@/api/types';
import { Section } from '@/components/Section';
import { EmptyState } from '@/components/EmptyState';
import { LatencyChip, ProviderChip, ReasonChip, ScoreBar } from '@/components/badges';
import { useTokens } from '@/lib/tokens';
import { ProviderBars } from '@/charts/ProviderBars';
import { gridDefaults, gridLocale, gridSx } from '@/lib/grid';
import type { TabProps } from './types';

export function ProvidersTab({ runId, run, goTab }: TabProps) {
  const { t, i18n } = useTranslation();
  const tk = useTokens();
  const q = useProviders(runId);
  const setOverride = useSetProviderOverride();
  const [edit, setEdit] = useState<{ p: ProviderStat; display: string; merged: string } | null>(null);
  const rows = useMemo(() => q.data?.providers ?? [], [q.data]);
  const canEdit = !!run.endpoint_id;
  const save = () => {
    if (!edit) return;
    setOverride.mutate(
      { endpointId: run.endpoint_id, raw_key: edit.p.key, display_name: edit.display.trim(), merged_into: edit.merged },
      { onSuccess: () => setEdit(null) },
    );
  };
  const columns: GridColDef<ProviderStat>[] = [
    {
      field: 'name',
      headerName: t('common.provider'),
      flex: 1,
      minWidth: 180,
      renderCell: (p) => <ProviderChip name={p.row.name} evidence={p.row.evidence} />,
    },
    { field: 'models', headerName: t('common.models'), width: 90 },
    { field: 'families', headerName: t('common.families'), width: 90 },
    { field: 'ok', headerName: t('runs.available'), width: 90 },
    {
      field: 'rate',
      headerName: t('run.providers.rate'),
      width: 160,
      renderCell: (p) => (
        <ScoreBar
          value={p.row.rate * 100}
          color={p.row.rate >= 0.8 ? tk.tierA : p.row.rate > 0 ? tk.tierB : tk.fail}
          digits={0}
          suffix="%"
        />
      ),
    },
    {
      field: 'median_s',
      headerName: t('common.median'),
      width: 110,
      renderCell: (p) => (p.row.median_s > 0 ? <LatencyChip seconds={p.row.median_s} /> : '–'),
    },
    {
      field: 'top_reason',
      headerName: t('run.providers.topReason'),
      width: 240,
      renderCell: (p) =>
        p.row.top_reason ? (
          <Stack direction="row" spacing={0.5} alignItems="center">
            <ReasonChip code={p.row.top_reason} />
            <span className="mono" style={{ fontSize: 12 }}>
              ×{p.row.top_reason_count}
            </span>
            {p.row.account_level_all && (
              <Tooltip title={t('run.providers.accountLevel')}>
                <Chip
                  size="small"
                  label="!"
                  sx={{ bgcolor: `${tk.fail}22`, color: tk.fail, fontWeight: 800, height: 20 }}
                />
              </Tooltip>
            )}
          </Stack>
        ) : (
          '–'
        ),
    },
    {
      field: 'actions',
      headerName: t('common.actions'),
      width: 170,
      sortable: false,
      renderCell: (p) => (
        <Stack direction="row" spacing={0.5} onClick={(e) => e.stopPropagation()}>
          <Button size="small" onClick={() => goTab('models', { provider: p.row.key })}>
            {t('common.models')}
          </Button>
          <Tooltip title={canEdit ? '' : t('newRun.inline')}>
            <span>
              <Button
                size="small"
                disabled={!canEdit}
                onClick={() => setEdit({ p: p.row, display: p.row.name, merged: '' })}
              >
                {t('common.edit')}
              </Button>
            </span>
          </Tooltip>
        </Stack>
      ),
    },
  ];
  if (q.error) return <EmptyState title={t('run.notScan')} description={String(q.error)} />;
  const suggestions = q.data?.suggestions ?? [];
  return (
    <Stack spacing={2}>
      {q.data && (
        <Typography variant="body2" color="text.secondary">
          {t('run.providers.rationale')}: <b>{q.data.field}</b> — {q.data.rationale}
        </Typography>
      )}
      {suggestions.length > 0 && (
        <Section title={t('run.providers.suggestions')}>
          <Stack spacing={1}>
            {suggestions.map((s, i) => (
              <Stack key={i} direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                {s.keys.map((k) => (
                  <Chip key={k} size="small" label={k} variant="outlined" />
                ))}
                <Typography variant="body2" color="text.secondary" sx={{ flex: 1 }}>
                  {s.evidence}
                </Typography>
                <Button
                  size="small"
                  variant="outlined"
                  disabled={!canEdit || setOverride.isPending}
                  onClick={() => {
                    for (const k of s.keys.slice(1))
                      setOverride.mutate({ endpointId: run.endpoint_id, raw_key: k, merged_into: s.keys[0] });
                  }}
                >
                  {t('run.providers.merge')} → {s.keys[0]}
                </Button>
              </Stack>
            ))}
          </Stack>
        </Section>
      )}
      <Section title={t('run.tabs.providers')}>
        <ProviderBars providers={rows} onPick={(k) => goTab('models', { provider: k })} />
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
          initialState={{
            pagination: { paginationModel: { pageSize: 50 } },
            sorting: { sortModel: [{ field: 'models', sort: 'desc' }] },
          }}
          rows={rows}
          getRowId={(r) => r.key}
          columns={columns}
          loading={q.isLoading}
          localeText={gridLocale(i18n.language)}
          autoHeight
          sx={{ ...gridSx, '& .MuiDataGrid-row:hover': { cursor: 'default' } }}
        />
      </Box>
      <Dialog
        open={!!edit}
        onClose={() => setEdit(null)}
        fullWidth
        maxWidth="xs"
        slotProps={{ paper: { sx: { borderRadius: 5 } } }}
      >
        <DialogTitle>
          {t('run.providers.rename')} · <span className="mono">{edit?.p.key}</span>
        </DialogTitle>
        <DialogContent>
          {edit && (
            <Stack spacing={2} sx={{ mt: 1 }}>
              <TextField
                label={t('settings.displayName')}
                value={edit.display}
                onChange={(e) => setEdit({ ...edit, display: e.target.value })}
                fullWidth
              />
              <FormControl fullWidth>
                <InputLabel>{t('run.providers.merge')}</InputLabel>
                <Select
                  label={t('run.providers.merge')}
                  value={edit.merged}
                  onChange={(e) => setEdit({ ...edit, merged: e.target.value })}
                >
                  <MenuItem value="">{t('common.none')}</MenuItem>
                  {rows
                    .filter((r) => r.key !== edit.p.key)
                    .map((r) => (
                      <MenuItem key={r.key} value={r.key}>
                        {r.name}{' '}
                        <span className="mono" style={{ opacity: 0.6, marginLeft: 8 }}>
                          {r.key}
                        </span>
                      </MenuItem>
                    ))}
                </Select>
              </FormControl>
              {setOverride.error && <Alert severity="error">{String(setOverride.error)}</Alert>}
              <Typography variant="caption" color="text.secondary">
                {t('settings.overridesHint')}
              </Typography>
            </Stack>
          )}
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setEdit(null)}>{t('common.cancel')}</Button>
          <Button variant="contained" disableElevation onClick={save} disabled={setOverride.isPending}>
            {t('common.save')}
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
