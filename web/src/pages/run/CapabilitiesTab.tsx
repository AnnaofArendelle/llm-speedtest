import { Box, Stack, Table, TableBody, TableCell, TableHead, TableRow, Tooltip, Typography } from '@mui/material';
import CheckCircleRounded from '@mui/icons-material/CheckCircleRounded';
import CancelRounded from '@mui/icons-material/CancelRounded';
import WarningAmberRounded from '@mui/icons-material/WarningAmberRounded';
import RemoveRounded from '@mui/icons-material/RemoveRounded';
import { useTranslation } from 'react-i18next';
import { useCapabilities, useProbes } from '@/api/hooks';
import { EmptyState } from '@/components/EmptyState';
import { Section } from '@/components/Section';
import { useTokens } from '@/lib/tokens';
import { fmtSeconds } from '@/lib/format';
import type { TabProps } from './types';

export default function CapabilitiesTab({ runId, onModel }: TabProps) {
  const { t, i18n } = useTranslation();
  const tk = useTokens();
  const q = useCapabilities(runId);
  const probes = useProbes();
  const zh = i18n.language.startsWith('zh');
  if (q.error) return <EmptyState title={t('common.empty')} description={String(q.error)} />;
  const m = q.data?.matrix;
  if (!m) return null;
  const models = m.models ?? [];
  if (models.length === 0) return <EmptyState title={t('common.empty')} />;
  const describe = (id: string) => probes.data?.find((p) => p.id === id)?.[zh ? 'zh' : 'en'] ?? id;
  const icon = (status: string) => {
    switch (status) {
      case 'supported':
        return <CheckCircleRounded sx={{ color: tk.tierA }} fontSize="small" />;
      case 'unsupported':
        return <CancelRounded sx={{ color: tk.fail, opacity: 0.8 }} fontSize="small" />;
      case 'error':
        return <WarningAmberRounded sx={{ color: tk.tierB }} fontSize="small" />;
      default:
        return <RemoveRounded sx={{ color: 'text.disabled' }} fontSize="small" />;
    }
  };
  const supportByProbe = (probe: string) => models.filter((id) => m.cells[id]?.[probe]?.status === 'supported').length;
  return (
    <Stack spacing={2}>
      <Typography variant="body2" color="text.secondary">
        {t('run.capabilities.intro')}
      </Typography>
      <Section noPad>
        <Box sx={{ overflowX: 'auto' }}>
          <Table
            size="small"
            stickyHeader
            sx={{
              '& td, & th': { px: 1, whiteSpace: 'nowrap' },
              '& th:first-of-type, & td:first-of-type': {
                position: 'sticky',
                left: 0,
                bgcolor: 'background.paper',
                zIndex: 1,
              },
            }}
          >
            <TableHead>
              <TableRow>
                <TableCell>{t('common.model')}</TableCell>
                <TableCell align="center">{t('run.capabilities.support')}</TableCell>
                {m.probes.map((p) => (
                  <TableCell key={p} align="center">
                    <Tooltip title={describe(p)}>
                      <Box>
                        <Typography variant="caption" sx={{ fontWeight: 700, display: 'block' }}>
                          {p}
                        </Typography>
                        <Typography variant="caption" color="text.secondary" className="mono">
                          {supportByProbe(p)}/{models.length}
                        </Typography>
                      </Box>
                    </Tooltip>
                  </TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {models.map((id) => (
                <TableRow key={id} hover>
                  <TableCell
                    className="mono"
                    sx={{ cursor: 'pointer', maxWidth: 280, overflow: 'hidden', textOverflow: 'ellipsis' }}
                    onClick={() => onModel(id)}
                    title={id}
                  >
                    {id}
                  </TableCell>
                  <TableCell align="center" className="mono" sx={{ fontWeight: 700 }}>
                    {m.support[id] ?? 0}/{m.probes.length}
                  </TableCell>
                  {m.probes.map((p) => {
                    const c = m.cells[id]?.[p];
                    return (
                      <TableCell key={p} align="center">
                        <Tooltip
                          title={
                            c ? (
                              <Box>
                                <Typography variant="subtitle2">
                                  {describe(p)} — {t(`run.capabilities.${c.status}`, { defaultValue: c.status })}
                                </Typography>
                                {c.evidence && (
                                  <Typography
                                    variant="body2"
                                    className="mono"
                                    sx={{ fontSize: 11, whiteSpace: 'pre-wrap' }}
                                  >
                                    {c.evidence}
                                  </Typography>
                                )}
                                {c.latency_s ? (
                                  <Typography variant="caption">{fmtSeconds(c.latency_s)}</Typography>
                                ) : null}
                              </Box>
                            ) : (
                              t('run.capabilities.skipped')
                            )
                          }
                        >
                          <Box sx={{ display: 'inline-flex' }}>{icon(c?.status ?? 'skipped')}</Box>
                        </Tooltip>
                      </TableCell>
                    );
                  })}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Box>
      </Section>
    </Stack>
  );
}
