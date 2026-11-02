import {
  Box,
  Chip,
  LinearProgress,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableRow,
  Tooltip,
  Typography,
} from '@mui/material';
import { useTranslation } from 'react-i18next';
import { useFamilyProfile, useModelProfile } from '@/api/hooks';
import type { Profile } from '@/api/types';
import { radius } from '@/app/theme';
import { RadarChart } from '@/charts/RadarChart';
import { arr, obj } from '@/lib/safe';
import { useTokens } from '@/lib/tokens';
import { ScoreBar } from './badges';

const STAGES = ['scan', 'latency', 'capability', 'quality', 'load'] as const;

export function ProfileView({ profile }: { profile: Profile }) {
  const { t, i18n } = useTranslation();
  const tk = useTokens();
  const zh = i18n.language.startsWith('zh');
  const dims = arr(profile.dimensions);
  const coverage = obj(profile.coverage);
  const measured = obj(profile.measured);
  return (
    <Box sx={{ p: 2, borderRadius: radius.block, bgcolor: 'background.paper' }}>
      <Stack direction="row" alignItems="baseline" justifyContent="space-between" spacing={1} sx={{ mb: 0.5 }}>
        <Typography variant="subtitle2">{t('profile.title')}</Typography>
        <Typography variant="caption" color="text.secondary">
          {t('profile.subtitle')}
        </Typography>
      </Stack>
      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 1, alignItems: 'center' }}>
        <RadarChart dims={dims} name={profile.display} height={250} />
        <Table size="small" sx={{ '& td': { px: 0.5, py: 0.4, borderBottom: 'none' } }} aria-label={t('profile.title')}>
          <TableBody>
            {dims.map((d) => (
              <TableRow key={d.key}>
                <TableCell
                  sx={{ whiteSpace: 'nowrap', color: d.covered ? 'text.primary' : 'text.disabled', fontWeight: 600 }}
                >
                  {zh ? d.label_zh : d.label_en}
                </TableCell>
                <TableCell>
                  {d.covered ? (
                    <Tooltip title={d.detail}>
                      <Box>
                        <ScoreBar
                          value={d.score}
                          width={72}
                          digits={0}
                          color={d.score >= 80 ? tk.tierA : d.score >= 50 ? tk.tierB : tk.fail}
                        />
                      </Box>
                    </Tooltip>
                  ) : (
                    <Typography variant="caption" color="text.disabled">
                      {t('profile.notCovered')}
                    </Typography>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
      <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap alignItems="center" sx={{ mt: 1 }}>
        <Typography variant="caption" color="text.secondary">
          {t('profile.coverage')}:
        </Typography>
        {STAGES.map((stage) => {
          const on = !!coverage[stage];
          const label =
            profile.kind === 'family' && on && profile.members > 1
              ? `${t(`kinds.${stage}`)} ${measured[stage] ?? 0}/${profile.members}`
              : t(`kinds.${stage}`);
          return (
            <Tooltip
              key={stage}
              title={
                on
                  ? profile.kind === 'family'
                    ? t('profile.measured', { n: measured[stage] ?? 0, total: profile.members })
                    : t('profile.covered')
                  : t('profile.notCovered')
              }
            >
              <Chip
                size="small"
                label={label}
                variant={on ? 'filled' : 'outlined'}
                sx={on ? { bgcolor: `${tk.tierA}22`, color: tk.tierA, fontWeight: 700 } : { color: 'text.disabled' }}
              />
            </Tooltip>
          );
        })}
      </Stack>
    </Box>
  );
}

export function ProfileCard({ runId, kind, id }: { runId: string; kind: 'model' | 'family'; id: string | undefined }) {
  const model = useModelProfile(kind === 'model' ? runId : undefined, kind === 'model' ? id : undefined);
  const family = useFamilyProfile(kind === 'family' ? runId : undefined, kind === 'family' ? id : undefined);
  const q = kind === 'model' ? model : family;
  if (q.isLoading) return <LinearProgress />;
  if (!q.data) return null;
  return <ProfileView profile={q.data} />;
}
