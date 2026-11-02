import { useEffect, useMemo, useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Collapse,
  IconButton,
  LinearProgress,
  Stack,
  Tooltip,
  Typography,
} from '@mui/material';
import SkipNextRounded from '@mui/icons-material/SkipNextRounded';
import ExpandMoreRounded from '@mui/icons-material/ExpandMoreRounded';
import CheckCircleRounded from '@mui/icons-material/CheckCircleRounded';
import ErrorOutlineRounded from '@mui/icons-material/ErrorOutlineRounded';
import { useTranslation } from 'react-i18next';
import type { InflightItem, Progress, Run, RunEvent } from '@/api/types';
import { radius } from '@/app/theme';
import { fmtDurationS } from '@/lib/format';
import { isLive } from '@/lib/run';
import { useTokens } from '@/lib/tokens';
import { ConfirmDialog } from './ConfirmDialog';
import { LiveFeed } from './LiveFeed';
import { StatusDot } from './badges';

const STAGE_ORDER = ['scan', 'latency', 'capability', 'quality', 'load'] as const;

function useNow(enabled: boolean) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!enabled) return;
    const h = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(h);
  }, [enabled]);
  return now;
}

function elapsed(since: string, now: number): string {
  const s = Math.max(0, Math.round((now - new Date(since).getTime()) / 1000));
  return s >= 60 ? `${Math.floor(s / 60)}m ${s % 60}s` : `${s}s`;
}

export interface ProgressCardProps {
  run: Run;
  childRuns: Run[];
  progress?: Progress;
  inflight: InflightItem[];
  feed: RunEvent[];
  logs: RunEvent[];
  onSkip: (model: string | undefined, all: boolean) => void;
  skipping?: boolean;
}

/** Live progress of a run: stage stepper, counters, in-flight requests with skip buttons and the result feed. */
export function ProgressCard({ run, childRuns, progress, inflight, feed, logs, onSkip, skipping }: ProgressCardProps) {
  const { t } = useTranslation();
  const tk = useTokens();
  const now = useNow(inflight.length > 0);
  const [confirmAll, setConfirmAll] = useState(false);
  const [showFeed, setShowFeed] = useState(true);
  const total = progress?.total ?? 0;
  const done = progress?.done ?? 0;
  const pct = total > 0 ? Math.min(100, (100 * done) / total) : 0;
  const currentStage = progress?.stage || (run.kind === 'suite' ? 'scan' : run.kind);
  const canSkip = currentStage !== 'load';

  const stages = useMemo(() => {
    const planned =
      run.kind === 'suite'
        ? STAGE_ORDER.filter((s) => s === 'scan' || childRuns.some((c) => c.kind === s) || s === currentStage)
        : [run.kind];
    return planned.map((stage) => {
      const child = childRuns.find((c) => c.kind === stage);
      let state: 'done' | 'running' | 'pending' | 'failed' | 'cancelled' = 'pending';
      if (child) {
        if (child.status === 'done') state = 'done';
        else if (isLive(child.status) || stage === currentStage) state = 'running';
        else if (child.status === 'failed') state = 'failed';
        else if (child.status === 'cancelled' || child.status === 'interrupted') state = 'cancelled';
      } else if (stage === currentStage) state = 'running';
      return { stage, state };
    });
  }, [run.kind, childRuns, currentStage]);

  const shown = inflight.slice(0, 8);
  return (
    <Card sx={{ mb: 2 }} data-testid="progress-card">
      <CardContent sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
          {stages.map(({ stage, state }, i) => {
            const color =
              state === 'done'
                ? tk.tierA
                : state === 'running'
                  ? tk.primary
                  : state === 'failed'
                    ? tk.fail
                    : state === 'cancelled'
                      ? tk.tierB
                      : tk.variant;
            return (
              <Stack key={stage} direction="row" spacing={1} alignItems="center">
                {i > 0 && <Box sx={{ width: 18, height: 2, bgcolor: 'divider', borderRadius: '1px' }} />}
                <Chip
                  size="small"
                  icon={
                    state === 'done' ? (
                      <CheckCircleRounded sx={{ fontSize: 16 }} />
                    ) : state === 'failed' ? (
                      <ErrorOutlineRounded sx={{ fontSize: 16 }} />
                    ) : (
                      <Box sx={{ display: 'flex', pl: 0.75 }}>
                        <StatusDot color="currentColor" pulse={state === 'running'} />
                      </Box>
                    )
                  }
                  label={t(`kinds.${stage}`)}
                  sx={{ bgcolor: `${color}1f`, color, fontWeight: 700, '& .MuiChip-icon': { color } }}
                  data-testid={`stage-${stage}-${state}`}
                />
              </Stack>
            );
          })}
        </Stack>
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: { xs: '1fr', md: 'minmax(260px, 1fr) minmax(300px, 1.3fr)' },
            gap: 2,
            alignItems: 'stretch',
          }}
        >
          <Stack direction="row" spacing={2} alignItems="center">
            <Box sx={{ position: 'relative', display: 'inline-flex', flexShrink: 0 }}>
              <CircularProgress
                variant="determinate"
                value={100}
                size={84}
                thickness={4}
                sx={{ color: 'action.hover', position: 'absolute' }}
              />
              <CircularProgress
                variant={total > 0 ? 'determinate' : 'indeterminate'}
                value={pct}
                size={84}
                thickness={4}
              />
              <Box
                sx={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
              >
                <Typography variant="subtitle1" className="mono" sx={{ fontWeight: 800 }}>
                  {total > 0 ? `${Math.round(pct)}%` : '…'}
                </Typography>
              </Box>
            </Box>
            <Stack spacing={0.5} sx={{ minWidth: 0 }}>
              <Typography variant="subtitle2" noWrap>
                {t('run.stage')}: {t(`kinds.${currentStage}`, { defaultValue: currentStage })}
                {progress?.rounds ? ` · ${t('run.round', { r: progress.round, n: progress.rounds })}` : ''}
              </Typography>
              <Typography variant="body2" className="mono">
                {done}/{total} · <span style={{ color: tk.tierA }}>✓ {progress?.ok ?? 0}</span> ·{' '}
                <span style={{ color: tk.fail }}>✗ {progress?.fail ?? 0}</span>
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {progress?.eta_s ? `${t('run.eta')} ${fmtDurationS(progress.eta_s)}` : t('run.live.etaPending')}
              </Typography>
              <LinearProgress
                variant={total > 0 ? 'determinate' : 'indeterminate'}
                value={pct}
                sx={{ height: 6, width: 180, maxWidth: '100%' }}
              />
            </Stack>
          </Stack>
          <Box sx={{ borderRadius: radius.block, bgcolor: 'background.default', p: 1.5, minHeight: 120 }}>
            <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1 }}>
              <Typography variant="overline" color="text.secondary" sx={{ flex: 1 }}>
                {t('run.live.inflight')} · {inflight.length}
              </Typography>
              {canSkip && inflight.length > 0 && (
                <Button
                  size="small"
                  color="warning"
                  startIcon={<SkipNextRounded />}
                  onClick={() => setConfirmAll(true)}
                  disabled={skipping}
                  data-testid="skip-all"
                >
                  {t('run.live.skipAll')}
                </Button>
              )}
            </Stack>
            {inflight.length === 0 ? (
              <Typography variant="body2" color="text.secondary">
                {t('run.live.none')}
              </Typography>
            ) : (
              <Stack spacing={0.25}>
                {shown.map((it) => (
                  <Stack key={it.id} direction="row" alignItems="center" spacing={1} data-testid="inflight-row">
                    <StatusDot color={tk.primary} pulse />
                    <Typography variant="body2" className="mono" noWrap sx={{ flex: 1, minWidth: 0 }} title={it.id}>
                      {it.id}
                    </Typography>
                    <Typography variant="caption" color="text.secondary" className="mono" sx={{ flexShrink: 0 }}>
                      {it.stage}
                      {it.round ? ` R${it.round}` : ''} · {elapsed(it.since, now)}
                    </Typography>
                    {canSkip && (
                      <Tooltip title={t('run.live.skip')}>
                        <span>
                          <IconButton
                            size="small"
                            color="warning"
                            onClick={() => onSkip(it.id, false)}
                            disabled={skipping}
                            aria-label={`${t('run.live.skip')} ${it.id}`}
                          >
                            <SkipNextRounded fontSize="small" />
                          </IconButton>
                        </span>
                      </Tooltip>
                    )}
                  </Stack>
                ))}
                {inflight.length > shown.length && (
                  <Typography variant="caption" color="text.secondary">
                    +{inflight.length - shown.length}
                  </Typography>
                )}
              </Stack>
            )}
          </Box>
        </Box>
        <Box>
          <Stack direction="row" alignItems="center" spacing={0.5}>
            <IconButton
              size="small"
              onClick={() => setShowFeed((v) => !v)}
              aria-label={t('run.liveFeed')}
              sx={{ transform: showFeed ? 'rotate(180deg)' : 'none', transition: 'transform .2s' }}
            >
              <ExpandMoreRounded fontSize="small" />
            </IconButton>
            <Typography variant="overline" color="text.secondary">
              {t('run.liveFeed')}
            </Typography>
            {logs.length > 0 && (
              <Typography variant="caption" color="text.secondary" noWrap sx={{ ml: 1, minWidth: 0 }}>
                {logs[logs.length - 1].message}
              </Typography>
            )}
          </Stack>
          <Collapse in={showFeed}>
            <LiveFeed events={feed} height={220} />
          </Collapse>
        </Box>
      </CardContent>
      <ConfirmDialog
        open={confirmAll}
        danger
        title={t('run.live.skipAll')}
        text={t('run.live.skipAllConfirm', { n: inflight.length })}
        confirmLabel={t('run.live.skipAll')}
        busy={skipping}
        onClose={() => setConfirmAll(false)}
        onConfirm={() => {
          onSkip(undefined, true);
          setConfirmAll(false);
        }}
      />
    </Card>
  );
}
