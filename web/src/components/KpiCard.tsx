import type { ReactNode } from 'react';
import { Box, Card, CardContent, Stack, Typography } from '@mui/material';
import { motion, useReducedMotion } from 'framer-motion';
import { CountUp } from './CountUp';

export function KpiCard({
  label,
  value,
  digits = 0,
  suffix,
  hint,
  icon,
  color,
  children,
  delay = 0,
}: {
  label: string;
  value?: number | string;
  digits?: number;
  suffix?: string;
  hint?: ReactNode;
  icon?: ReactNode;
  color?: string;
  children?: ReactNode;
  /** Entrance delay in seconds (stagger within a grid). */
  delay?: number;
}) {
  const reduce = useReducedMotion();
  return (
    <motion.div
      initial={reduce ? false : { opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, delay: reduce ? 0 : delay, ease: 'easeOut' }}
      style={{ height: '100%' }}
    >
      <Card sx={{ height: '100%' }}>
        <CardContent
          sx={{ display: 'flex', flexDirection: 'column', gap: 0.5, py: 2.25, '&:last-child': { pb: 2.25 } }}
        >
          <Stack direction="row" alignItems="center" justifyContent="space-between" spacing={1}>
            <Typography variant="overline" color="text.secondary" sx={{ letterSpacing: 1, lineHeight: 1.4 }} noWrap>
              {label}
            </Typography>
            {icon && <Box sx={{ color: color ?? 'primary.main', display: 'flex', opacity: 0.9 }}>{icon}</Box>}
          </Stack>
          {(value !== undefined || !children) && (
            <Typography
              variant="h4"
              className="mono"
              sx={{ color, fontVariantNumeric: 'tabular-nums', lineHeight: 1.15 }}
            >
              {typeof value === 'number' ? <CountUp value={value} digits={digits} /> : (value ?? '–')}
              {suffix && (
                <Typography component="span" variant="h6" color="text.secondary" sx={{ ml: 0.5 }}>
                  {suffix}
                </Typography>
              )}
            </Typography>
          )}
          {hint && (
            <Typography variant="body2" color="text.secondary" component="div">
              {hint}
            </Typography>
          )}
          {children}
        </CardContent>
      </Card>
    </motion.div>
  );
}
