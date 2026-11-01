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
import { useFamily } from '@/api/hooks';
import { useReasonLookup } from '@/lib/reasons';
import { tierColor } from '@/app/theme';
import { LatencyChip, MiniStat, ReasonChip, ScoreBar, TierBadge } from './badges';
import { useTokens } from '@/lib/tokens';

export function FamilyDrawer({
  runId,
  familyKey,
  onClose,
  onModel,
}: {
  runId: string;
  familyKey?: string;
  onClose: () => void;
  onModel?: (id: string) => void;
}) {
  const { t } = useTranslation();
  const tk = useTokens();
  const lookup = useReasonLookup();
  const q = useFamily(runId, familyKey);
  const f = q.data?.family;
  const members = q.data?.members ?? [];
  return (
    <Drawer
      anchor="right"
      open={!!familyKey}
      onClose={onClose}
      slotProps={{ paper: { sx: { width: { xs: '100vw', sm: 640 }, bgcolor: 'background.default' } } }}
    >
      <Box sx={{ p: { xs: 2, sm: 3 }, display: 'flex', flexDirection: 'column', gap: 2 }}>
        <Stack direction="row" alignItems="flex-start" spacing={1}>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography variant="overline" color="text.secondary">
              {t('common.family')}
            </Typography>
            <Typography variant="h6" sx={{ lineHeight: 1.25 }}>
              {f?.display ?? familyKey}
            </Typography>
            <Typography variant="caption" color="text.secondary" className="mono">
              {t('run.families.key')}: {familyKey}
            </Typography>
          </Box>
          <IconButton onClick={onClose} aria-label={t('common.close')}>
            <CloseRounded />
          </IconButton>
        </Stack>
        {q.isLoading && <LinearProgress />}
        {f && (
          <>
            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
              <TierBadge tier={f.tier} />
              {f.rank_composite ? (
                <Chip size="small" variant="outlined" label={`#${f.rank_composite} ${t('run.tabs.ranking')}`} />
              ) : null}
              {f.rank_stability ? (
                <Chip size="small" variant="outlined" label={`#${f.rank_stability} ${t('run.tabs.stability')}`} />
              ) : null}
              {f.median_s > 0 && <LatencyChip seconds={f.median_s} />}
            </Stack>
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: 'repeat(3, 1fr)',
                gap: 2,
                p: 2,
                borderRadius: 4,
                bgcolor: 'background.paper',
              }}
            >
              <MiniStat
                label={t('common.score')}
                value={<ScoreBar value={f.score} color={tierColor(tk, f.tier)} width={64} />}
                mono={false}
              />
              <MiniStat
                label={t('run.ranking.stabilityScore')}
                value={<ScoreBar value={f.s_stability} width={64} digits={0} />}
                mono={false}
              />
              <MiniStat
                label={t('run.ranking.breadth')}
                value={<ScoreBar value={f.s_breadth} width={64} digits={0} />}
                mono={false}
              />
              <MiniStat
                label={t('run.ranking.speed')}
                value={<ScoreBar value={f.s_speed} width={64} digits={0} />}
                mono={false}
              />
              <MiniStat label={t('run.ranking.availableMembers')} value={`${f.ok}/${f.n}`} />
              <MiniStat
                label={t('run.ranking.providersOk')}
                value={`${f.n_prov_ok}/${Object.keys(f.providers).length}`}
              />
            </Box>
            <Box>
              <Typography variant="subtitle2" sx={{ mb: 0.75 }}>
                {t('common.providers')}
              </Typography>
              <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap>
                {Object.entries(f.providers)
                  .sort((a, b) => b[1] - a[1])
                  .map(([p, n]) => {
                    const ok = f.ok_providers[p] ?? 0;
                    const color = ok === n ? tk.tierA : ok > 0 ? tk.tierB : tk.fail;
                    return (
                      <Chip
                        key={p}
                        size="small"
                        label={`${p} ${ok}/${n}`}
                        sx={{ bgcolor: `${color}1f`, color, fontWeight: 600 }}
                      />
                    );
                  })}
              </Stack>
            </Box>
            {Object.keys(f.reasons).length > 0 && (
              <Box>
                <Typography variant="subtitle2" sx={{ mb: 0.75 }}>
                  {t('run.ranking.topReasons')}
                </Typography>
                <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap>
                  {Object.entries(f.reasons)
                    .sort((a, b) => b[1] - a[1])
                    .map(([r, n]) => (
                      <Stack key={r} direction="row" spacing={0.5} alignItems="center">
                        <ReasonChip code={r} />
                        <Typography variant="caption" className="mono">
                          ×{n}
                        </Typography>
                      </Stack>
                    ))}
                </Stack>
              </Box>
            )}
            <Typography variant="body2" color="text.secondary">
              {t('run.families.sources')}:{' '}
              {Object.entries(f.sources)
                .map(([s, n]) => `${s} ×${n}`)
                .join(' · ') || t('common.none')}
            </Typography>
            <Box>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>
                {t('run.ranking.members')} ({members.length})
              </Typography>
              <Table size="small" sx={{ '& td, & th': { px: 1 } }}>
                <TableHead>
                  <TableRow>
                    <TableCell>{t('common.model')}</TableCell>
                    <TableCell>{t('common.provider')}</TableCell>
                    <TableCell>{t('common.status')}</TableCell>
                    <TableCell>{t('common.median')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {members.map((m) => (
                    <TableRow
                      key={m.id}
                      hover
                      onClick={onModel ? () => onModel(m.id) : undefined}
                      sx={{ cursor: onModel ? 'pointer' : 'default' }}
                    >
                      <TableCell
                        className="mono"
                        sx={{ maxWidth: 260, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                        title={m.id}
                      >
                        {m.id}
                      </TableCell>
                      <TableCell sx={{ whiteSpace: 'nowrap' }}>{m.provider_name}</TableCell>
                      <TableCell>
                        {m.status === 'ok' ? <TierBadge tier={m.tier} /> : <ReasonChip code={m.reason} />}
                      </TableCell>
                      <TableCell>
                        {m.status === 'ok' ? (
                          <LatencyChip seconds={m.median_s} />
                        ) : (
                          <span title={lookup.name(m.reason)}>–</span>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </Box>
          </>
        )}
      </Box>
    </Drawer>
  );
}
