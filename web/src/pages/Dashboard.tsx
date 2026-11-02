import { useMemo } from 'react';
import { Link as RouterLink, useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  Chip,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material';
import BoltRounded from '@mui/icons-material/BoltRounded';
import HubRounded from '@mui/icons-material/HubRounded';
import HistoryRounded from '@mui/icons-material/HistoryRounded';
import { useEndpoints, useRuns, useSummary } from '@/api/hooks';
import { KpiCard } from '@/components/KpiCard';
import { Section } from '@/components/Section';
import { EmptyState } from '@/components/EmptyState';
import { PageHeader } from '@/components/PageHeader';
import { KindChip, LatencyChip, ProviderChip, StatusChip, TierBadge } from '@/components/badges';
import { useTokens } from '@/lib/tokens';
import { LatencyHistogram } from '@/charts/LatencyHistogram';
import { Donut } from '@/charts/Donut';
import { FamilyBars } from '@/charts/FamilyBars';
import { fmtPct, fmtRelative, fmtSeconds } from '@/lib/format';
import { useReasonLookup } from '@/lib/reasons';
import { runTitle, scanCounts } from '@/lib/run';

export default function Dashboard() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const tk = useTokens();
  const lookup = useReasonLookup();
  const runs = useRuns(undefined, 30);
  const endpoints = useEndpoints();
  const topLevel = useMemo(() => (runs.data?.runs ?? []).filter((r) => !r.parent_id), [runs.data]);
  const latest = useMemo(
    () => topLevel.find((r) => (r.kind === 'suite' || r.kind === 'scan') && r.status === 'done'),
    [topLevel],
  );
  const summary = useSummary(latest?.id);
  const doc = summary.data;
  const s = doc?.summary;
  const reasonData = useMemo(() => {
    if (!s?.reasons) return [];
    const entries = Object.entries(s.reasons).sort((a, b) => b[1] - a[1]);
    const top = entries.slice(0, 8).map(([k, v]) => ({ key: k, name: lookup.name(k), value: v }));
    const rest = entries.slice(8).reduce((acc, [, v]) => acc + v, 0);
    if (rest > 0) top.push({ key: '', name: t('common.more'), value: rest });
    return top;
  }, [s, lookup, t]);
  const topReason = reasonData[0];

  if (runs.isSuccess && topLevel.length === 0) {
    return (
      <>
        <PageHeader title={t('dashboard.title')} subtitle={t('dashboard.subtitle')} />
        <EmptyState
          icon={<BoltRounded fontSize="inherit" />}
          title={t('dashboard.noRuns')}
          action={
            <Stack direction="row" spacing={1}>
              <Button variant="outlined" component={RouterLink} to="/endpoints" startIcon={<HubRounded />}>
                {t('endpoints.add')}
              </Button>
              <Button variant="contained" disableElevation component={RouterLink} to="/new" startIcon={<BoltRounded />}>
                {t('dashboard.start')}
              </Button>
            </Stack>
          }
        />
      </>
    );
  }

  return (
    <>
      <PageHeader
        title={t('dashboard.title')}
        subtitle={
          latest ? (
            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
              <span>
                {t('dashboard.latestRun')}: <b>{runTitle(latest)}</b>
              </span>
              <KindChip kind={latest.kind} />
              <Chip size="small" variant="outlined" label={latest.preset} />
              <Typography variant="body2" color="text.secondary">
                {fmtRelative(latest.finished_at || latest.created_at)}
              </Typography>
              <Button size="small" component={RouterLink} to={`/runs/${latest.id}`}>
                {t('common.open')}
              </Button>
            </Stack>
          ) : (
            t('dashboard.subtitle')
          )
        }
        actions={
          <Button variant="contained" disableElevation component={RouterLink} to="/new" startIcon={<BoltRounded />}>
            {t('dashboard.start')}
          </Button>
        }
      />
      {s && (
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr 1fr', md: 'repeat(4, 1fr)' }, gap: 2, mb: 2 }}>
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
            hint={`${s.providers} ${t('common.providers')}`}
          />
          <KpiCard
            label={t('dashboard.kpi.median')}
            value={fmtSeconds(s.latency_median_s)}
            hint={`${t('dashboard.kpi.p90')} ${fmtSeconds(s.latency_p90_s)} · ${t('common.min')} ${fmtSeconds(s.latency_min_s)}`}
          />
          <KpiCard label={t('dashboard.kpi.tiers')}>
            <Stack direction="row" spacing={1} sx={{ mt: 0.5 }}>
              {(['A', 'B', 'C'] as const).map((tier) => (
                <Chip
                  key={tier}
                  size="small"
                  label={`${tier} ${s.tier_counts?.[tier] ?? 0}`}
                  sx={{
                    bgcolor: `${tier === 'A' ? tk.tierA : tier === 'B' ? tk.tierB : tk.tierC}22`,
                    color: tier === 'A' ? tk.tierA : tier === 'B' ? tk.tierB : tk.tierC,
                    fontWeight: 800,
                  }}
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
            value={topReason ? topReason.value : '–'}
            hint={topReason?.name ?? t('common.none')}
          />
          <KpiCard
            label={t('common.providers')}
            value={s.providers}
            hint={`${t('run.overview.providerField')}: ${doc?.provider_field ?? '–'}`}
          />
        </Box>
      )}
      {s && (
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', lg: '1fr 1fr' }, gap: 2, mb: 2 }}>
          <Section title={t('run.overview.latencyDist')} subtitle={t('run.overview.caveat')}>
            <LatencyHistogram buckets={s.buckets ?? []} />
          </Section>
          <Section title={t('run.overview.reasons')}>
            <Donut
              data={reasonData}
              onPick={(k) => void navigate(`/runs/${latest?.id}?tab=models&status=fail${k ? `&reason=${k}` : ''}`)}
            />
          </Section>
        </Box>
      )}
      {doc && ((doc.top_families ?? []).length > 0 || (doc.top_models ?? []).length > 0) && (
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', lg: '1fr 1fr' }, gap: 2, mb: 2 }}>
          <Section
            title={t('run.overview.topFamilies')}
            action={
              <Button size="small" component={RouterLink} to={`/runs/${latest?.id}?tab=ranking`}>
                {t('dashboard.viewAll')}
              </Button>
            }
          >
            <FamilyBars
              families={(doc.top_families ?? []).slice(0, 10)}
              onPick={(k) => void navigate(`/runs/${latest?.id}?tab=ranking&family=${encodeURIComponent(k)}`)}
            />
          </Section>
          <Section
            title={t('run.overview.topModels')}
            noPad
            action={
              <Button size="small" component={RouterLink} to={`/runs/${latest?.id}?tab=models`}>
                {t('dashboard.viewAll')}
              </Button>
            }
          >
            <Box sx={{ overflowX: 'auto' }}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>#</TableCell>
                    <TableCell>{t('common.model')}</TableCell>
                    <TableCell>{t('common.provider')}</TableCell>
                    <TableCell>{t('common.stability')}</TableCell>
                    <TableCell>{t('common.median')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {(doc.top_models ?? []).slice(0, 10).length > 0 ? (
                    (doc.top_models ?? []).slice(0, 10).map((m, i) => (
                      <TableRow
                        key={m.id}
                        hover
                        sx={{ cursor: 'pointer' }}
                        onClick={() =>
                          void navigate(`/runs/${latest?.id}?tab=models&model=${encodeURIComponent(m.id)}`)
                        }
                      >
                        <TableCell className="mono">{i + 1}</TableCell>
                        <TableCell
                          className="mono"
                          sx={{ maxWidth: 260, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                          title={m.id}
                        >
                          {m.id}
                        </TableCell>
                        <TableCell>
                          <ProviderChip name={m.provider_name} />
                        </TableCell>
                        <TableCell>
                          <TierBadge tier={m.tier} />
                        </TableCell>
                        <TableCell>
                          <LatencyChip seconds={m.median_s} />
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={5} align="center">
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
      )}
      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', lg: '3fr 2fr' }, gap: 2 }}>
        <Section
          title={t('dashboard.recentRuns')}
          action={
            <Button size="small" component={RouterLink} to="/runs" startIcon={<HistoryRounded />}>
              {t('dashboard.viewAll')}
            </Button>
          }
        >
          <Stack spacing={1}>
            {topLevel.slice(0, 6).map((r) => {
              const c = scanCounts(r);
              return (
                <Card key={r.id} variant="outlined" sx={{ boxShadow: 'none' }}>
                  <CardActionArea component={RouterLink} to={`/runs/${r.id}`} sx={{ p: 1.5 }}>
                    <Stack direction="row" spacing={1.5} alignItems="center" flexWrap="wrap" useFlexGap>
                      <StatusChip status={r.status} />
                      <KindChip kind={r.kind} />
                      <Typography variant="body2" sx={{ fontWeight: 600, flex: 1, minWidth: 120 }} noWrap>
                        {runTitle(r)}
                      </Typography>
                      {c && (
                        <Typography variant="body2" color="text.secondary" className="mono">
                          {c.ok}/{c.models}
                        </Typography>
                      )}
                      <Typography variant="caption" color="text.secondary">
                        {fmtRelative(r.created_at)}
                      </Typography>
                    </Stack>
                  </CardActionArea>
                </Card>
              );
            })}
          </Stack>
        </Section>
        <Section
          title={t('dashboard.endpoints')}
          action={
            <Button size="small" component={RouterLink} to="/endpoints" startIcon={<HubRounded />}>
              {t('dashboard.viewAll')}
            </Button>
          }
        >
          <Stack spacing={1}>
            {(endpoints.data ?? []).slice(0, 6).map((e) => (
              <Card key={e.id} variant="outlined" sx={{ boxShadow: 'none' }}>
                <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      <Typography variant="body2" sx={{ fontWeight: 600 }} noWrap>
                        {e.name}
                      </Typography>
                      <Typography variant="caption" color="text.secondary" className="mono" noWrap component="div">
                        {e.base_url}
                      </Typography>
                    </Box>
                    {e.dialect && <Chip size="small" variant="outlined" label={e.dialect} />}
                    <Button size="small" variant="outlined" component={RouterLink} to={`/new?endpoint=${e.id}`}>
                      {t('endpoints.testNow')}
                    </Button>
                  </Stack>
                </CardContent>
              </Card>
            ))}
            {endpoints.isSuccess && endpoints.data.length === 0 && (
              <EmptyState
                compact
                title={t('endpoints.none')}
                action={
                  <Button variant="outlined" component={RouterLink} to="/endpoints">
                    {t('endpoints.add')}
                  </Button>
                }
              />
            )}
          </Stack>
        </Section>
      </Box>
    </>
  );
}
