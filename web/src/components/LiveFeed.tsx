import { Box, Stack, Typography } from '@mui/material';
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import type { RunEvent } from '@/api/types';
import { fmtSeconds } from '@/lib/format';
import { LatencyChip, ReasonChip } from './badges';

function eventKey(ev: RunEvent): string {
  if (ev.seq) return String(ev.seq);
  return `${ev.at}-${ev.attempt?.model_id ?? ev.sample?.model_id ?? ''}-${ev.attempt?.round ?? ev.sample?.index ?? ''}`;
}

export function LiveFeed({ events, height = 360 }: { events: RunEvent[]; height?: number }) {
  const { t } = useTranslation();
  const reduce = useReducedMotion();
  if (!events.length) {
    return (
      <Typography color="text.secondary" variant="body2" sx={{ p: 2 }}>
        {t('run.waitingEvents')}
      </Typography>
    );
  }
  return (
    <Box sx={{ maxHeight: height, overflow: 'auto', px: 1 }}>
      <AnimatePresence initial={false}>
        {events.map((ev) => {
          const a = ev.attempt;
          const s = ev.sample;
          const id = a?.model_id ?? s?.model_id ?? '';
          return (
            <motion.div
              key={eventKey(ev)}
              layout={!reduce}
              initial={reduce ? false : { opacity: 0, x: -14 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0 }}
              transition={{ type: 'spring', stiffness: 420, damping: 32 }}
            >
              <Stack
                direction="row"
                alignItems="center"
                spacing={1}
                sx={{ py: 0.6, borderBottom: '1px solid', borderColor: 'divider' }}
              >
                <Typography variant="caption" color="text.secondary" className="mono" sx={{ width: 34, flexShrink: 0 }}>
                  {a ? `R${a.round}` : s ? `#${s.index + 1}` : ''}
                </Typography>
                <Typography variant="body2" className="mono" noWrap sx={{ flex: 1, minWidth: 0 }} title={id}>
                  {id}
                </Typography>
                {a && (a.ok ? <LatencyChip seconds={a.latency / 1e9} /> : <ReasonChip code={a.reason} />)}
                {s &&
                  (s.ok ? (
                    <Typography variant="caption" className="mono" color="text.secondary" noWrap>
                      TTFT {fmtSeconds(s.ttft / 1e9)} · {s.tok_per_s.toFixed(1)} tok/s
                    </Typography>
                  ) : (
                    <ReasonChip code={s.reason ?? 'unknown'} />
                  ))}
              </Stack>
            </motion.div>
          );
        })}
      </AnimatePresence>
    </Box>
  );
}
