import { lazy, Suspense, useMemo, useState } from 'react';
import { Link as RouterLink, useParams, useSearchParams } from 'react-router';
import { useTranslation } from 'react-i18next';
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  LinearProgress,
  Skeleton,
  Stack,
  Tab,
  Tabs,
  Typography,
} from '@mui/material';
import StopCircleOutlined from '@mui/icons-material/StopCircleOutlined';
import DownloadRounded from '@mui/icons-material/DownloadRounded';
import CompareArrowsRounded from '@mui/icons-material/CompareArrowsRounded';
import { useCancelRun, useRun } from '@/api/hooks';
import { useRunEvents } from '@/api/sse';
import type { Progress, Run } from '@/api/types';
import { PageHeader } from '@/components/PageHeader';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { KindChip, StatusChip } from '@/components/badges';
import { LiveFeed } from '@/components/LiveFeed';
import { ModelDrawer } from '@/components/ModelDrawer';
import { FamilyDrawer } from '@/components/FamilyDrawer';
import { Section } from '@/components/Section';
import { durationBetween, fmtDate, fmtDurationS } from '@/lib/format';
import { isLive, runTitle } from '@/lib/run';
import { OverviewTab } from './run/OverviewTab';
import { RankingTab } from './run/RankingTab';
import { ModelsTab } from './run/ModelsTab';
import { ProvidersTab } from './run/ProvidersTab';
import type { TabProps } from './run/types';

const FailuresTab = lazy(() => import('./run/FailuresTab'));
const FamiliesTab = lazy(() => import('./run/FamiliesTab'));
const LatencyTab = lazy(() => import('./run/LatencyTab'));
const LoadTab = lazy(() => import('./run/LoadTab'));
const CapabilitiesTab = lazy(() => import('./run/CapabilitiesTab'));
const QualityTab = lazy(() => import('./run/QualityTab'));
const ExportTab = lazy(() => import('./run/ExportTab'));

const SCAN_TABS = ['overview', 'ranking', 'stability', 'models', 'providers', 'failures', 'families'] as const;

function ProgressCard({
  run,
  progress,
  children,
  feed,
}: {
  run: Run;
  progress: Progress | undefined;
  children: Run[];
  feed: Parameters<typeof LiveFeed>[0]['events'];
}) {
  const { t } = useTranslation();
  const total = progress?.total ?? 0;
  const done = progress?.done ?? 0;
  const pct = total > 0 ? Math.min(100, (100 * done) / total) : 0;
  const stage = progress?.stage || run.status;
  return (
    <Card sx={{ mb: 2 }}>
      <CardContent>
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} alignItems={{ md: 'center' }}>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap sx={{ mb: 1 }}>
              <Chip
                size="small"
                color="primary"
                label={`${t('run.stage')}: ${t(`kinds.${stage}`, { defaultValue: stage })}`}
              />
              {progress?.rounds ? (
                <Chip
                  size="small"
                  variant="outlined"
                  label={t('run.round', { r: progress.round, n: progress.rounds })}
                />
              ) : null}
              <Typography variant="body2" className="mono">
                {done}/{total}
              </Typography>
              <Typography variant="body2" color="success.main" className="mono">
                ✓ {progress?.ok ?? 0}
              </Typography>
              <Typography variant="body2" color="error.main" className="mono">
                ✗ {progress?.fail ?? 0}
              </Typography>
              {progress?.eta_s ? (
                <Typography variant="body2" color="text.secondary">
                  {t('run.eta')} {fmtDurationS(progress.eta_s)}
                </Typography>
              ) : null}
            </Stack>
            <LinearProgress
              variant={total > 0 ? 'determinate' : 'indeterminate'}
              value={pct}
              sx={{ height: 8, borderRadius: 4 }}
            />
            {children.length > 0 && (
              <Stack direction="row" spacing={0.75} sx={{ mt: 1.5 }} flexWrap="wrap" useFlexGap>
                {children.map((c) => (
                  <Chip
                    key={c.id}
                    size="small"
                    variant="outlined"
                    label={`${t(`kinds.${c.kind}`)} · ${t(`common.${c.status}`, { defaultValue: c.status })}`}
                  />
                ))}
              </Stack>
            )}
          </Box>
          <Box sx={{ width: { xs: '100%', md: 460 }, flexShrink: 0 }}>
            <Typography variant="overline" color="text.secondary">
              {t('run.liveFeed')}
            </Typography>
            <LiveFeed events={feed} height={220} />
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}

export default function RunDetail() {
  const { id = '' } = useParams();
  const { t } = useTranslation();
  const [params, setParams] = useSearchParams();
  const runQ = useRun(id);
  const run = runQ.data?.run;
  const children = useMemo(() => runQ.data?.children ?? [], [runQ.data]);
  const running = runQ.data?.running ?? false;
  const live = useRunEvents(id, running);
  const cancel = useCancelRun();
  const [confirmCancel, setConfirmCancel] = useState(false);
  const [modelId, setModelId] = useState<string | undefined>(params.get('model') ?? undefined);
  const [familyKey, setFamilyKey] = useState<string | undefined>(params.get('family') ?? undefined);

  const kinds = useMemo(() => new Set([run?.kind, ...children.map((c) => c.kind)]), [run, children]);
  const tabs = useMemo(() => {
    if (!run) return [];
    const list: string[] = [];
    if (run.kind === 'scan' || run.kind === 'suite') list.push(...SCAN_TABS);
    if (kinds.has('latency')) list.push('latency');
    if (kinds.has('load')) list.push('load');
    if (kinds.has('capability')) list.push('capabilities');
    if (kinds.has('quality')) list.push('quality');
    list.push('export');
    return list;
  }, [run, kinds]);
  const requested = params.get('tab') ?? '';
  const tab = tabs.includes(requested) ? requested : (tabs[0] ?? 'export');
  const goTab = (next: string, extra?: Record<string, string>) =>
    setParams(
      (prev) => {
        const keep = prev.get('tab');
        for (const k of [...prev.keys()]) if (k !== 'tab') prev.delete(k);
        prev.set('tab', next);
        if (extra) for (const [k, v] of Object.entries(extra)) prev.set(k, v);
        if (keep === next && !extra) prev.set('tab', next);
        return prev;
      },
      { replace: true },
    );

  if (runQ.error) return <Alert severity="error">{String(runQ.error)}</Alert>;
  if (!run) return <Skeleton variant="rounded" height={240} sx={{ borderRadius: 5 }} />;
  const progress: Progress | undefined =
    (live.latest && live.latest.kind !== 'done' ? live.latest : undefined) ?? run.progress ?? undefined;
  const tabProps: TabProps = { runId: id, run, onModel: setModelId, onFamily: setFamilyKey, goTab };
  const showProgress = running || isLive(run.status);

  return (
    <>
      <PageHeader
        back="/runs"
        title={
          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
            <span>{runTitle(run)}</span>
            <StatusChip status={running ? 'running' : run.status} size="medium" />
          </Stack>
        }
        subtitle={
          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
            <KindChip kind={run.kind} />
            <Chip size="small" variant="outlined" label={run.preset} />
            {run.dialect && <Chip size="small" variant="outlined" label={run.dialect} />}
            <Typography variant="body2" color="text.secondary" className="mono">
              {run.base_url}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {fmtDate(run.started_at || run.created_at)}
              {run.finished_at && run.started_at ? ` · ${durationBetween(run.started_at, run.finished_at)}` : ''}
            </Typography>
            {run.note && <Typography variant="body2">{run.note}</Typography>}
          </Stack>
        }
        actions={
          <>
            {showProgress && (
              <Button
                variant="outlined"
                color="warning"
                startIcon={<StopCircleOutlined />}
                onClick={() => setConfirmCancel(true)}
              >
                {t('runs.cancel')}
              </Button>
            )}
            {(run.kind === 'scan' || run.kind === 'suite') && run.status === 'done' && (
              <Button
                variant="outlined"
                startIcon={<CompareArrowsRounded />}
                component={RouterLink}
                to={`/compare?a=${run.id}`}
              >
                {t('nav.compare')}
              </Button>
            )}
            <Button
              variant="contained"
              disableElevation
              startIcon={<DownloadRounded />}
              onClick={() => goTab('export')}
            >
              {t('common.export')}
            </Button>
          </>
        }
      />
      {run.error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {run.error}
        </Alert>
      )}
      {live.error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {live.error}
        </Alert>
      )}
      {showProgress && <ProgressCard run={run} progress={progress} children={children} feed={live.feed} />}
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs
          value={tab}
          onChange={(_e, v: string) => goTab(v)}
          variant="scrollable"
          scrollButtons="auto"
          allowScrollButtonsMobile
        >
          {tabs.map((k) => (
            <Tab key={k} value={k} label={t(`run.tabs.${k}`)} />
          ))}
        </Tabs>
      </Box>
      <Suspense fallback={<LinearProgress />}>
        {tab === 'overview' && <OverviewTab {...tabProps} />}
        {tab === 'ranking' && <RankingTab {...tabProps} order="composite" />}
        {tab === 'stability' && <RankingTab {...tabProps} order="stability" />}
        {tab === 'models' && <ModelsTab {...tabProps} />}
        {tab === 'providers' && <ProvidersTab {...tabProps} />}
        {tab === 'failures' && <FailuresTab {...tabProps} />}
        {tab === 'families' && <FamiliesTab {...tabProps} />}
        {tab === 'latency' && <LatencyTab {...tabProps} />}
        {tab === 'load' && <LoadTab {...tabProps} />}
        {tab === 'capabilities' && <CapabilitiesTab {...tabProps} />}
        {tab === 'quality' && <QualityTab {...tabProps} />}
        {tab === 'export' && <ExportTab {...tabProps} />}
      </Suspense>
      {!showProgress && live.rounds.length > 0 && (
        <Section title={t('run.progress')} sx={{ mt: 2 }}>
          <Stack spacing={0.5}>
            {live.rounds.map((r, i) => (
              <Typography key={i} variant="body2" className="mono" color="text.secondary">
                {r.message}
              </Typography>
            ))}
          </Stack>
        </Section>
      )}
      <ModelDrawer
        runId={id}
        modelId={modelId}
        onClose={() => setModelId(undefined)}
        onFamily={(k) => {
          setModelId(undefined);
          setFamilyKey(k);
        }}
      />
      <FamilyDrawer
        runId={id}
        familyKey={familyKey}
        onClose={() => setFamilyKey(undefined)}
        onModel={(m) => {
          setFamilyKey(undefined);
          setModelId(m);
        }}
      />
      <ConfirmDialog
        open={confirmCancel}
        title={t('runs.cancel')}
        text={t('run.cancelConfirm')}
        danger
        confirmLabel={t('runs.cancel')}
        busy={cancel.isPending}
        onClose={() => setConfirmCancel(false)}
        onConfirm={() => cancel.mutate(id, { onSuccess: () => setConfirmCancel(false) })}
      />
    </>
  );
}
