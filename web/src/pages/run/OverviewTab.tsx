import { useMemo } from 'react';
import { Box, Chip, Stack, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';
import { useTranslation } from 'react-i18next';
import { useSummary } from '@/api/hooks';
import { KpiCard } from '@/components/KpiCard';
import { Section } from '@/components/Section';
import { EmptyState } from '@/components/EmptyState';
import { LatencyChip, ProviderChip, TierBadge } from '@/components/badges';
import { useTokens } from '@/lib/tokens';
import { LatencyHistogram } from '@/charts/LatencyHistogram';
import { Donut } from '@/charts/Donut';
import { FamilyBars } from '@/charts/FamilyBars';
import { fmtPct, fmtSeconds } from '@/lib/format';
import { useReasonLookup } from '@/lib/reasons';
import type { TabProps } from './types';

export function OverviewTab({ runId, onModel, onFamily, goTab }: TabProps) {
  const { t } = useTranslation();
  const tk = useTokens();
  const lookup = useReasonLookup();
  const q = useSummary(runId);
  const doc = q.data;
  const s = doc?.summary;
  const reasonData = useMemo(() => {
    if (!s?.reasons) return [];
    const entries = Object.entries(s.reasons).sort((a, b) => b[1] - a[1]);
    const top = entries.slice(0, 9).map(([k, v]) => ({ key: k, name: lookup.name(k), value: v }));
    const rest = entries.slice(9).reduce((acc, [, v]) => acc + v, 0);
    if (rest > 0) top.push({ key: '', name: t('common.more'), value: rest });
    return top;
  }, [s, lookup, t]);
  if (q.error) return <EmptyState title={t('run.notScan')} description={String(q.error)} />;
  if (!doc || !s || !doc.scoring) return null;
  const tierColors = { A: tk.tierA, B: tk.tierB, C: tk.tierC } as const;
  return (
    <Stack spacing={2}>
      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr 1fr', md: 'repeat(4, 1fr)' }, gap: 2 }}>
        <KpiCard
          label={t('dashboard.kpi.models')}
          value={s.models}
          hint={t('run.overview.kpiHint', { attempts: s.attempts, rounds: s.rounds })}
        />
        <KpiCard
          label={t('dashboard.kpi.available')}
          value={s.ok}
          color={tk.tierA}
          hint={`${t('dashboard.kpi.availableRate')} ${fmtPct(s.ok, s.models)}`}
        />
        <KpiCard
          label={t('dashboard.kpi.families')}
          value={s.families_ok}
          suffix={`/ ${s.families}`}
          hint={`${s.families_unavailable} ${t('tier.none')}`}
        />
        <KpiCard
          label={t('dashboard.kpi.median')}
          value={fmtSeconds(s.latency_median_s)}
          hint={`${t('dashboard.kpi.p90')} ${fmtSeconds(s.latency_p90_s)} · ${t('common.min')} ${fmtSeconds(s.latency_min_s)} · ${t('common.max')} ${fmtSeconds(s.latency_max_s)}`}
        />
        <KpiCard label={t('dashboard.kpi.tiers')}>
          <Stack direction="row" spacing={1} sx={{ mt: 0.5 }} flexWrap="wrap" useFlexGap>
            {(['A', 'B', 'C'] as const).map((tier) => (
              <Chip
                key={tier}
                size="small"
                label={`${tier} ${s.tier_counts?.[tier] ?? 0}`}
                onClick={() => goTab('models', { tier, status: 'ok' })}
                sx={{ bgcolor: `${tierColors[tier]}22`, color: tierColors[tier], fontWeight: 800 }}
              />
            ))}
          </Stack>
        </KpiCard>
        <KpiCard
          label={t('dashboard.kpi.providerLevel')}
          value={s.provider_level_fail}
          color={tk.fail}
          hint={`${fmtPct(s.provider_level_fail, s.fail)} ${t('tier.none')}`}
        />
        <KpiCard
          label={t('run.overview.topReason')}
          value={reasonData[0]?.value ?? '–'}
          hint={reasonData[0]?.name ?? t('common.none')}
        />
        <KpiCard
          label={t('common.providers')}
          value={s.providers}
          hint={`${t('run.overview.providerField')}: ${doc.provider_field}`}
        />
      </Box>
      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', lg: '1fr 1fr' }, gap: 2 }}>
        <Section title={t('run.overview.latencyDist')} subtitle={t('run.overview.caveat')}>
          <LatencyHistogram buckets={s.buckets ?? []} />
        </Section>
        <Section title={t('run.overview.reasons')} subtitle={`${s.fail} ${t('tier.none')}`}>
          <Donut
            data={reasonData}
            onPick={(k) => goTab('models', k ? { status: 'fail', reason: k } : { status: 'fail' })}
          />
        </Section>
      </Box>
      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', lg: '1fr 1fr' }, gap: 2 }}>
        <Section
          title={t('run.overview.topFamilies')}
          subtitle={t('run.ranking.intro', {
            ws: doc.scoring.w_stability,
            wb: doc.scoring.w_breadth,
            wsp: doc.scoring.w_speed,
          })}
        >
          <FamilyBars families={doc.top_families ?? []} onPick={onFamily} />
        </Section>
        <Section title={t('run.overview.topModels')} noPad>
          <Box sx={{ overflowX: 'auto' }}>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>#</TableCell>
                  <TableCell>{t('common.model')}</TableCell>
                  <TableCell>{t('common.provider')}</TableCell>
                  <TableCell>{t('common.stability')}</TableCell>
                  <TableCell>{t('common.median')}</TableCell>
                  <TableCell>
                    {t('common.min')}–{t('common.max')}
                  </TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {(doc.top_models ?? []).length > 0 ? (
                  (doc.top_models ?? []).map((m, i) => (
                    <TableRow key={m.id} hover sx={{ cursor: 'pointer' }} onClick={() => onModel(m.id)}>
                      <TableCell className="mono">{i + 1}</TableCell>
                      <TableCell
                        className="mono"
                        sx={{ maxWidth: 240, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                        title={m.id}
                      >
                        {m.id}
                      </TableCell>
                      <TableCell>
                        <ProviderChip name={m.provider_name} />
                      </TableCell>
                      <TableCell>
                        <Stack direction="row" spacing={0.5} alignItems="center">
                          <TierBadge tier={m.tier} />
                          <Typography variant="caption" className="mono" color="text.secondary">
                            {m.ok}/{m.attempts}
                          </Typography>
                        </Stack>
                      </TableCell>
                      <TableCell>
                        <LatencyChip seconds={m.median_s} />
                      </TableCell>
                      <TableCell className="mono" sx={{ whiteSpace: 'nowrap' }}>
                        {fmtSeconds(m.min_s)} – {fmtSeconds(m.max_s)}
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={6} align="center">
                      <Typography variant="body2" color="text.secondary" sx={{ py: 2 }}>
                        {t('common.none')}
                      </Typography>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </Box>
        </Section>
      </Box>
    </Stack>
  );
}
