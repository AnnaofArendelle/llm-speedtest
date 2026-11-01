import { useState } from 'react';
import {
  Box,
  Button,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material';
import { useTranslation } from 'react-i18next';
import { useOverrides, useSetFamilyOverride, useSetProviderOverride } from '@/api/hooks';
import { EmptyState } from './EmptyState';

export function OverridesPanel({ endpointId }: { endpointId: string }) {
  const { t } = useTranslation();
  const q = useOverrides(endpointId);
  const setProv = useSetProviderOverride();
  const setFam = useSetFamilyOverride();
  const [p, setP] = useState({ raw_key: '', display_name: '', merged_into: '' });
  const [f, setF] = useState({ model_id: '', family_key: '', canon: '' });
  const providers = q.data?.providers ?? [];
  const families = q.data?.families ?? [];
  return (
    <Stack spacing={3}>
      <Typography variant="body2" color="text.secondary">
        {t('settings.overridesHint')}
      </Typography>
      <Box>
        <Typography variant="subtitle2" sx={{ mb: 1 }}>
          {t('settings.providerOverrides')} ({providers.length})
        </Typography>
        {providers.length > 0 ? (
          <Table size="small" sx={{ mb: 1.5 }}>
            <TableHead>
              <TableRow>
                <TableCell>{t('settings.rawKey')}</TableCell>
                <TableCell>{t('settings.displayName')}</TableCell>
                <TableCell>{t('settings.mergedInto')}</TableCell>
                <TableCell>{t('settings.source')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {providers.map((r) => (
                <TableRow
                  key={r.raw_key}
                  hover
                  onClick={() => setP({ raw_key: r.raw_key, display_name: r.display_name, merged_into: r.merged_into })}
                  sx={{ cursor: 'pointer' }}
                >
                  <TableCell className="mono">{r.raw_key}</TableCell>
                  <TableCell>{r.display_name || '–'}</TableCell>
                  <TableCell className="mono">{r.merged_into || '–'}</TableCell>
                  <TableCell>{r.source}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <Box sx={{ mb: 1.5 }}>
            <EmptyState compact title={t('settings.noOverrides')} />
          </Box>
        )}
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems="flex-start">
          <TextField
            size="small"
            label={t('settings.rawKey')}
            value={p.raw_key}
            onChange={(e) => setP({ ...p, raw_key: e.target.value })}
            required
            sx={{ minWidth: 160 }}
          />
          <TextField
            size="small"
            label={t('settings.displayName')}
            value={p.display_name}
            onChange={(e) => setP({ ...p, display_name: e.target.value })}
            sx={{ minWidth: 160 }}
          />
          <TextField
            size="small"
            label={t('settings.mergedInto')}
            value={p.merged_into}
            onChange={(e) => setP({ ...p, merged_into: e.target.value })}
            sx={{ minWidth: 160 }}
          />
          <Button
            variant="contained"
            disableElevation
            disabled={!p.raw_key || setProv.isPending}
            onClick={() =>
              setProv.mutate(
                {
                  endpointId,
                  raw_key: p.raw_key.trim(),
                  display_name: p.display_name.trim(),
                  merged_into: p.merged_into.trim(),
                },
                { onSuccess: () => setP({ raw_key: '', display_name: '', merged_into: '' }) },
              )
            }
          >
            {t('common.save')}
          </Button>
        </Stack>
      </Box>
      <Box>
        <Typography variant="subtitle2" sx={{ mb: 1 }}>
          {t('settings.familyOverrides')} ({families.length})
        </Typography>
        {families.length > 0 ? (
          <Table size="small" sx={{ mb: 1.5 }}>
            <TableHead>
              <TableRow>
                <TableCell>{t('settings.modelId')}</TableCell>
                <TableCell>{t('settings.familyKey')}</TableCell>
                <TableCell>{t('settings.canon')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {families.map((r) => (
                <TableRow
                  key={r.model_id}
                  hover
                  onClick={() => setF({ model_id: r.model_id, family_key: r.family_key, canon: r.canon })}
                  sx={{ cursor: 'pointer' }}
                >
                  <TableCell className="mono">{r.model_id}</TableCell>
                  <TableCell className="mono">{r.family_key}</TableCell>
                  <TableCell>{r.canon || '–'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <Box sx={{ mb: 1.5 }}>
            <EmptyState compact title={t('settings.noOverrides')} />
          </Box>
        )}
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems="flex-start">
          <TextField
            size="small"
            label={t('settings.modelId')}
            value={f.model_id}
            onChange={(e) => setF({ ...f, model_id: e.target.value })}
            required
            sx={{ minWidth: 200 }}
          />
          <TextField
            size="small"
            label={t('settings.familyKey')}
            value={f.family_key}
            onChange={(e) => setF({ ...f, family_key: e.target.value })}
            required
            sx={{ minWidth: 160 }}
            helperText="deepseek|flash|4"
          />
          <TextField
            size="small"
            label={t('settings.canon')}
            value={f.canon}
            onChange={(e) => setF({ ...f, canon: e.target.value })}
            sx={{ minWidth: 160 }}
          />
          <Button
            variant="contained"
            disableElevation
            disabled={!f.model_id || !f.family_key || setFam.isPending}
            onClick={() =>
              setFam.mutate(
                { endpointId, model_id: f.model_id.trim(), family_key: f.family_key.trim(), canon: f.canon.trim() },
                { onSuccess: () => setF({ model_id: '', family_key: '', canon: '' }) },
              )
            }
          >
            {t('common.save')}
          </Button>
        </Stack>
      </Box>
    </Stack>
  );
}
