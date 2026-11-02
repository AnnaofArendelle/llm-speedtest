import {
  Box,
  Chip,
  Drawer,
  IconButton,
  LinearProgress,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material';
import CloseRounded from '@mui/icons-material/CloseRounded';
import { useTranslation } from 'react-i18next';
import { radius } from '@/app/theme';
import { useModel } from '@/api/hooks';
import { fmtContext, fmtDate, fmtSeconds } from '@/lib/format';
import { useReasonLookup } from '@/lib/reasons';
import { CopyButton } from './CopyButton';
import { ErrorBoundary } from './ErrorBoundary';
import { ProfileCard } from './ProfileCard';
import { LatencyChip, MiniStat, ProviderChip, ReasonChip, TierBadge } from './badges';
import { useTokens } from '@/lib/tokens';

export function ModelDrawer({
  runId,
  modelId,
  onClose,
  onFamily,
}: {
  runId: string;
  modelId?: string;
  onClose: () => void;
  onFamily?: (key: string) => void;
}) {
  const { t } = useTranslation();
  const tk = useTokens();
  const lookup = useReasonLookup();
  const q = useModel(runId, modelId);
  const m = q.data?.model;
  const history = q.data?.history ?? [];
  return (
    <Drawer
      anchor="right"
      open={!!modelId}
      onClose={onClose}
      slotProps={{ paper: { sx: { width: { xs: '100vw', sm: 600 }, bgcolor: 'background.default' } } }}
    >
      <Box sx={{ p: { xs: 2, sm: 3 }, display: 'flex', flexDirection: 'column', gap: 2 }}>
        <Stack direction="row" alignItems="flex-start" spacing={1}>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography variant="overline" color="text.secondary">
              {t('common.model')}
            </Typography>
            <Typography variant="h6" className="mono" sx={{ wordBreak: 'break-all', lineHeight: 1.25 }}>
              {modelId}
            </Typography>
          </Box>
          <CopyButton text={modelId ?? ''} />
          <IconButton onClick={onClose} aria-label={t('common.close')}>
            <CloseRounded />
          </IconButton>
        </Stack>
        {q.isLoading && <LinearProgress />}
        {q.error && (
          <Typography color="error" variant="body2">
            {String(q.error)}
          </Typography>
        )}
        {m && (
          <>
            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap alignItems="center">
              {m.status === 'ok' ? (
                <Chip
                  size="small"
                  label={t('runs.available')}
                  sx={{ bgcolor: `${tk.tierA}22`, color: tk.tierA, fontWeight: 700 }}
                />
              ) : (
                <ReasonChip code={m.reason} />
              )}
              <TierBadge tier={m.tier} />
              {m.skipped ? (
                <Chip
                  size="small"
                  label={t('run.models.skipped', { n: m.skipped })}
                  sx={{ bgcolor: `${tk.tierB}22`, color: tk.tierB, fontWeight: 700 }}
                />
              ) : null}
              <ProviderChip name={m.provider_name} />
              <Chip
                size="small"
                label={m.family_canon}
                onClick={onFamily ? () => onFamily(m.family_key) : undefined}
                variant="outlined"
                sx={{ fontWeight: 600 }}
              />
              {m.context_length ? (
                <Chip
                  size="small"
                  variant="outlined"
                  label={`${t('run.models.context')} ${fmtContext(m.context_length)}`}
                />
              ) : null}
            </Stack>
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: 'repeat(3, 1fr)',
                gap: 2,
                p: 2,
                borderRadius: radius.block,
                bgcolor: 'background.paper',
              }}
            >
              <MiniStat label={t('common.median')} value={fmtSeconds(m.median_s)} />
              <MiniStat label={t('common.min')} value={fmtSeconds(m.min_s)} />
              <MiniStat label={t('common.max')} value={fmtSeconds(m.max_s)} />
              <MiniStat label={t('run.models.jitter')} value={m.jitter > 0 ? `${m.jitter.toFixed(2)}×` : '–'} />
              <MiniStat label={t('common.attempts')} value={`${m.ok}/${m.attempts}`} />
              <MiniStat label={t('common.stability')} value={`${(m.rate * 100).toFixed(0)}%`} />
            </Box>
            <ErrorBoundary label={t('profile.title')} resetKeys={[m.id]}>
              <ProfileCard runId={runId} kind="model" id={m.id} />
            </ErrorBoundary>
            {m.served_model && (
              <Typography variant="body2">
                {t('run.models.served')}: <span className="mono">{m.served_model}</span>
              </Typography>
            )}
            {m.flaky && Object.keys(m.flaky).length > 0 && (
              <Typography variant="body2" color="warning.main">
                {t('run.models.flaky', {
                  list: Object.entries(m.flaky)
                    .map(([k, v]) => `${lookup.name(k)} ×${v}`)
                    .join(', '),
                })}
              </Typography>
            )}
            {m.status !== 'ok' && (
              <Box>
                <Typography variant="subtitle2" sx={{ mb: 0.5 }}>
                  {t('run.models.message')}{' '}
                  <Typography component="span" variant="caption" color="text.secondary" className="mono">
                    HTTP {m.http || '–'}
                    {m.inner ? ` / ${m.inner}` : ''}
                  </Typography>
                </Typography>
                <Box
                  component="pre"
                  className="mono"
                  sx={{
                    m: 0,
                    p: 1.5,
                    borderRadius: radius.control,
                    bgcolor: 'background.paper',
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                    fontSize: 12,
                    maxHeight: 220,
                    overflow: 'auto',
                  }}
                >
                  {m.message || lookup.desc(m.reason) || m.reason}
                </Box>
                {lookup.fix(m.reason) && (
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                    → {lookup.fix(m.reason)}
                  </Typography>
                )}
              </Box>
            )}
            <Typography variant="body2" color="text.secondary">
              {t('run.models.familySource')}: <b>{m.family_source}</b>
              {m.family_evidence ? ` · ${m.family_evidence}` : ''}
            </Typography>
            {history.length > 0 && (
              <Box>
                <Typography variant="subtitle2" sx={{ mb: 1 }}>
                  {t('run.models.history')}
                </Typography>
                <Table size="small" sx={{ '& td, & th': { px: 1, whiteSpace: 'nowrap' } }}>
                  <TableHead>
                    <TableRow>
                      <TableCell>R</TableCell>
                      <TableCell>{t('common.status')}</TableCell>
                      <TableCell>{t('common.latency')}</TableCell>
                      <TableCell>HTTP</TableCell>
                      <TableCell>{t('run.models.served')}</TableCell>
                      <TableCell>{t('common.createdAt')}</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {history.map((h, i) => (
                      <TableRow key={i} hover>
                        <TableCell className="mono">{h.round}</TableCell>
                        <TableCell>
                          {h.ok ? (
                            <Chip
                              size="small"
                              label="OK"
                              sx={{ bgcolor: `${tk.tierA}22`, color: tk.tierA, fontWeight: 700 }}
                            />
                          ) : (
                            <ReasonChip code={h.reason} />
                          )}
                        </TableCell>
                        <TableCell>
                          <LatencyChip seconds={h.latency / 1e9} />
                        </TableCell>
                        <TableCell className="mono">
                          {h.http || '–'}
                          {h.inner ? `/${h.inner}` : ''}
                        </TableCell>
                        <TableCell
                          className="mono"
                          sx={{ maxWidth: 160, overflow: 'hidden', textOverflow: 'ellipsis' }}
                        >
                          {h.served_model || '–'}
                        </TableCell>
                        <TableCell className="mono">{fmtDate(h.at)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </Box>
            )}
          </>
        )}
      </Box>
    </Drawer>
  );
}
