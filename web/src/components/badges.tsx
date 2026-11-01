import type { ReactNode } from 'react';
import { Box, Chip, Tooltip, Typography, useTheme } from '@mui/material';
import { alpha } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';
import { latencyColor, tierColor } from '@/app/theme';
import { useTokens } from '@/lib/tokens';
import { fmtSeconds } from '@/lib/format';
import { useReasonLookup } from '@/lib/reasons';

export function StatusDot({ color, pulse, size = 8 }: { color: string; pulse?: boolean; size?: number }) {
  return (
    <Box
      component="span"
      className={pulse ? 'lst-pulse' : undefined}
      sx={{ display: 'inline-block', width: size, height: size, borderRadius: '50%', bgcolor: color, flexShrink: 0 }}
    />
  );
}

export function TierBadge({ tier, size = 'small' }: { tier?: string; size?: 'small' | 'medium' }) {
  const tk = useTokens();
  const { t } = useTranslation();
  const color = tierColor(tk, tier);
  return (
    <Tooltip title={t(`tier.${tier || 'none'}`)}>
      <Chip
        size={size}
        label={tier || '✕'}
        sx={{ bgcolor: alpha(color, 0.16), color, fontWeight: 800, minWidth: 36, '& .MuiChip-label': { px: 1 } }}
      />
    </Tooltip>
  );
}

export function LatencyChip({
  seconds,
  size = 'small',
  prefix,
}: {
  seconds: number;
  size?: 'small' | 'medium';
  prefix?: string;
}) {
  const tk = useTokens();
  const color = latencyColor(tk, seconds);
  return (
    <Chip
      size={size}
      className="mono"
      label={`${prefix ?? ''}${fmtSeconds(seconds)}`}
      sx={{ bgcolor: alpha(color, 0.14), color, fontWeight: 700 }}
    />
  );
}

export function ReasonChip({
  code,
  size = 'small',
  showCode,
}: {
  code: string;
  size?: 'small' | 'medium';
  showCode?: boolean;
}) {
  const tk = useTokens();
  const theme = useTheme();
  const lookup = useReasonLookup();
  const info = lookup.info(code);
  const color = !info
    ? theme.palette.text.secondary
    : info.provider_level
      ? tk.fail
      : info.transient
        ? tk.tierB
        : tk.tierC;
  const title = info ? (
    <Box>
      <Typography variant="subtitle2">
        {lookup.name(code)} <span className="mono">({code})</span>
      </Typography>
      <Typography variant="body2">{lookup.desc(code)}</Typography>
      {lookup.fix(code) && (
        <Typography variant="body2" sx={{ mt: 0.5, opacity: 0.85 }}>
          → {lookup.fix(code)}
        </Typography>
      )}
    </Box>
  ) : (
    code
  );
  return (
    <Tooltip title={title}>
      <Chip
        size={size}
        label={showCode ? code : lookup.name(code)}
        sx={{ bgcolor: alpha(color, 0.14), color, fontWeight: 600, maxWidth: 240 }}
      />
    </Tooltip>
  );
}

export function ProviderChip({
  name,
  evidence,
  size = 'small',
  onClick,
}: {
  name: string;
  evidence?: { field: string; value: string }[];
  size?: 'small' | 'medium';
  onClick?: () => void;
}) {
  const title = evidence?.length ? (
    <Box>
      {evidence.map((e, i) => (
        <Typography key={i} variant="body2">
          <b>{e.field}</b> = <span className="mono">{e.value}</span>
        </Typography>
      ))}
    </Box>
  ) : (
    name
  );
  return (
    <Tooltip title={title}>
      <Chip size={size} label={name} variant="outlined" onClick={onClick} sx={{ maxWidth: 200, fontWeight: 600 }} />
    </Tooltip>
  );
}

export function KindChip({ kind, size = 'small' }: { kind: string; size?: 'small' | 'medium' }) {
  const { t } = useTranslation();
  return <Chip size={size} variant="outlined" label={t(`kinds.${kind}`, { defaultValue: kind })} />;
}

export function StatusChip({ status, size = 'small' }: { status: string; size?: 'small' | 'medium' }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const tk = useTokens();
  const live = status === 'running' || status === 'pending' || status === 'discovering';
  const color = live
    ? theme.palette.primary.main
    : status === 'done'
      ? tk.tierA
      : status === 'failed'
        ? tk.fail
        : tk.tierB;
  return (
    <Chip
      size={size}
      icon={
        <Box sx={{ display: 'flex', pl: 0.75 }}>
          <StatusDot color={color} pulse={live} />
        </Box>
      }
      label={t(`common.${status}`, { defaultValue: status })}
      sx={{ bgcolor: alpha(color, 0.12), color, fontWeight: 700 }}
    />
  );
}

export function ScoreBar({
  value,
  max = 100,
  color,
  width = 96,
  digits = 1,
  suffix,
}: {
  value: number;
  max?: number;
  color?: string;
  width?: number | string;
  digits?: number;
  suffix?: string;
}) {
  const pct = Math.max(0, Math.min(100, (value / max) * 100));
  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, minWidth: 0 }}>
      <Box sx={{ width, height: 6, borderRadius: 3, bgcolor: 'action.hover', overflow: 'hidden', flexShrink: 0 }}>
        <Box
          sx={{
            width: `${pct}%`,
            height: '100%',
            bgcolor: color ?? 'primary.main',
            borderRadius: 3,
            transition: 'width .4s ease',
          }}
        />
      </Box>
      <Typography variant="body2" className="mono" sx={{ fontVariantNumeric: 'tabular-nums' }}>
        {value.toFixed(digits)}
        {suffix}
      </Typography>
    </Box>
  );
}

export function MiniStat({ label, value, mono = true }: { label: string; value: ReactNode; mono?: boolean }) {
  return (
    <Box>
      <Typography variant="caption" color="text.secondary" sx={{ display: 'block', lineHeight: 1.2 }}>
        {label}
      </Typography>
      <Typography
        variant="body1"
        className={mono ? 'mono' : undefined}
        sx={{ fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}
      >
        {value}
      </Typography>
    </Box>
  );
}
