import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Box,
  Chip,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material';
import ExpandMoreRounded from '@mui/icons-material/ExpandMoreRounded';
import { useTranslation } from 'react-i18next';
import { useFailures } from '@/api/hooks';
import type { ModelStat } from '@/api/types';
import { EmptyState } from '@/components/EmptyState';
import { CopyButton } from '@/components/CopyButton';
import { ReasonChip } from '@/components/badges';
import { useTokens } from '@/lib/tokens';
import { fmtSeconds, truncate } from '@/lib/format';
import { useReasonLookup } from '@/lib/reasons';
import type { TabProps } from './types';

function ProviderBlock({
  provider,
  models,
  onModel,
}: {
  provider: string;
  models: ModelStat[];
  onModel: (id: string) => void;
}) {
  const { t } = useTranslation();
  const messages = new Set(models.map((m) => m.message ?? ''));
  const same = messages.size === 1 && models.length > 1;
  return (
    <Box sx={{ mb: 2 }}>
      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 0.5 }}>
        <Typography variant="subtitle2">{provider}</Typography>
        <Chip size="small" label={models.length} />
        {same && (
          <Typography variant="caption" color="text.secondary">
            {t('run.failures.sameMessage', { n: models.length })}
          </Typography>
        )}
      </Stack>
      {same && (
        <Box
          component="pre"
          className="mono"
          sx={{
            m: 0,
            mb: 1,
            p: 1.25,
            borderRadius: 3,
            bgcolor: 'background.default',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            fontSize: 12,
            position: 'relative',
          }}
        >
          {models[0].message || '–'}
          <Box sx={{ position: 'absolute', top: 4, right: 4 }}>
            <CopyButton text={models[0].message ?? ''} />
          </Box>
        </Box>
      )}
      <Box sx={{ overflowX: 'auto' }}>
        <Table size="small" sx={{ '& td, & th': { px: 1 } }}>
          <TableHead>
            <TableRow>
              <TableCell>{t('common.model')}</TableCell>
              <TableCell>HTTP</TableCell>
              <TableCell>{t('common.latency')}</TableCell>
              {!same && <TableCell>{t('run.models.message')}</TableCell>}
            </TableRow>
          </TableHead>
          <TableBody>
            {models.map((m) => (
              <TableRow key={m.id} hover sx={{ cursor: 'pointer' }} onClick={() => onModel(m.id)}>
                <TableCell className="mono" sx={{ whiteSpace: 'nowrap' }}>
                  {m.id}
                </TableCell>
                <TableCell className="mono">
                  {m.http || '–'}
                  {m.inner ? `/${m.inner}` : ''}
                </TableCell>
                <TableCell className="mono">{m.fail_latency_s ? fmtSeconds(m.fail_latency_s) : '–'}</TableCell>
                {!same && (
                  <TableCell className="mono" sx={{ fontSize: 12, maxWidth: 520 }} title={m.message}>
                    {truncate(m.message, 160) || '–'}
                  </TableCell>
                )}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
    </Box>
  );
}

export default function FailuresTab({ runId, onModel }: TabProps) {
  const { t } = useTranslation();
  const tk = useTokens();
  const lookup = useReasonLookup();
  const q = useFailures(runId);
  if (q.error) return <EmptyState title={t('run.notScan')} description={String(q.error)} />;
  const groups = q.data?.groups ?? [];
  if (q.isSuccess && groups.length === 0) return <EmptyState title={t('common.empty')} />;
  return (
    <Stack spacing={1.5}>
      <Typography variant="body2" color="text.secondary">
        {t('run.failures.intro')}
      </Typography>
      {groups.map((g) => (
        <Accordion
          key={g.reason}
          disableGutters
          sx={{ borderRadius: 5, '&:before': { display: 'none' }, overflow: 'hidden' }}
        >
          <AccordionSummary expandIcon={<ExpandMoreRounded />}>
            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap sx={{ pr: 2 }}>
              <ReasonChip code={g.reason} size="medium" />
              <Chip size="small" label={g.count} sx={{ fontWeight: 800 }} />
              <Chip
                size="small"
                variant="outlined"
                label={g.info.provider_level ? t('run.failures.level.provider') : t('run.failures.level.model')}
                sx={{ color: g.info.provider_level ? tk.fail : undefined }}
              />
              <Typography variant="body2" color="text.secondary">
                {lookup.desc(g.reason)}
              </Typography>
            </Stack>
          </AccordionSummary>
          <AccordionDetails>
            {lookup.fix(g.reason) && (
              <Typography variant="body2" sx={{ mb: 2 }}>
                <b>{t('run.failures.fix')}:</b> {lookup.fix(g.reason)}
              </Typography>
            )}
            {Object.entries(g.by_provider)
              .sort((a, b) => b[1].length - a[1].length)
              .map(([p, models]) => (
                <ProviderBlock key={p} provider={p} models={models} onModel={onModel} />
              ))}
          </AccordionDetails>
        </Accordion>
      ))}
    </Stack>
  );
}
