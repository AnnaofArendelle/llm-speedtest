import { useState } from 'react';
import { Link as RouterLink } from 'react-router';
import { useTranslation } from 'react-i18next';
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  IconButton,
  InputLabel,
  LinearProgress,
  MenuItem,
  Select,
  Stack,
  Switch,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Tooltip,
  Typography,
} from '@mui/material';
import AddRounded from '@mui/icons-material/AddRounded';
import EditRounded from '@mui/icons-material/EditRounded';
import DeleteOutlineRounded from '@mui/icons-material/DeleteOutlineRounded';
import HubRounded from '@mui/icons-material/HubRounded';
import BoltRounded from '@mui/icons-material/BoltRounded';
import TuneRounded from '@mui/icons-material/TuneRounded';
import {
  useCheckEndpoint,
  useDeleteEndpoint,
  useDialects,
  useDiscoverEndpoint,
  useEndpoints,
  useSaveEndpoint,
  type EndpointInput,
} from '@/api/hooks';
import type { CheckResult, DiscoverResult, Endpoint } from '@/api/types';
import { PageHeader } from '@/components/PageHeader';
import { EmptyState } from '@/components/EmptyState';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { OverridesPanel } from '@/components/OverridesPanel';
import { useTokens } from '@/lib/tokens';
import { fmtSeconds } from '@/lib/format';
import { arr } from '@/lib/safe';

function parseHeaders(text: string): Record<string, string> {
  const out: Record<string, string> = {};
  for (const line of text.split('\n')) {
    const i = line.indexOf('=');
    if (i <= 0) continue;
    out[line.slice(0, i).trim()] = line.slice(i + 1).trim();
  }
  return out;
}

export function CheckResultView({ r }: { r: CheckResult }) {
  const { t } = useTranslation();
  const tk = useTokens();
  return (
    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap alignItems="center">
      <Chip
        size="small"
        label={r.reachable ? t('endpoints.reachable') : t('endpoints.unreachable')}
        sx={{
          bgcolor: `${r.reachable ? tk.tierA : tk.fail}22`,
          color: r.reachable ? tk.tierA : tk.fail,
          fontWeight: 700,
        }}
      />
      {r.reachable && (
        <Chip
          size="small"
          label={r.auth_ok ? t('endpoints.authOk') : t('endpoints.authFail')}
          sx={{
            bgcolor: `${r.auth_ok ? tk.tierA : tk.fail}22`,
            color: r.auth_ok ? tk.tierA : tk.fail,
            fontWeight: 700,
          }}
        />
      )}
      {r.auth_ok && (
        <Chip size="small" variant="outlined" label={t('endpoints.modelsFound', { n: r.models, o: r.owners })} />
      )}
      {r.dialect?.name && (
        <Chip
          size="small"
          variant="outlined"
          label={`${t('endpoints.detected', { d: r.dialect.name })}${r.dialect.version ? ` ${r.dialect.version}` : ''}`}
        />
      )}
      {r.latency_s > 0 && <Chip size="small" variant="outlined" className="mono" label={fmtSeconds(r.latency_s)} />}
      {r.error && (
        <Typography variant="caption" color="error" sx={{ wordBreak: 'break-all' }}>
          {r.error}
        </Typography>
      )}
      {r.suggested_base_url && (
        <Typography variant="caption" color="warning.main" className="mono">
          → {r.suggested_base_url}
        </Typography>
      )}
    </Stack>
  );
}

function EndpointDialog({ open, initial, onClose }: { open: boolean; initial?: Endpoint; onClose: () => void }) {
  const { t } = useTranslation();
  const dialects = useDialects();
  const save = useSaveEndpoint();
  const [form, setForm] = useState<EndpointInput & { headersText: string }>(() => ({
    id: initial?.id,
    name: initial?.name ?? '',
    base_url: initial?.base_url ?? '',
    api_key: '',
    clear_key: false,
    dialect: initial?.dialect ?? '',
    insecure: initial?.insecure ?? false,
    proxy: initial?.proxy ?? '',
    headersText: Object.entries(initial?.headers ?? {})
      .map(([k, v]) => `${k}=${v}`)
      .join('\n'),
  }));
  const submit = () => {
    const { headersText, ...rest } = form;
    save.mutate(
      { ...rest, headers: parseHeaders(headersText), base_url: rest.base_url.trim(), name: rest.name.trim() },
      { onSuccess: onClose },
    );
  };
  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>{initial ? t('endpoints.edit') : t('endpoints.add')}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <TextField
            label={t('endpoints.name')}
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            fullWidth
          />
          <TextField
            label={t('endpoints.baseUrl')}
            value={form.base_url}
            onChange={(e) => setForm({ ...form, base_url: e.target.value })}
            fullWidth
            required
            placeholder="https://api.example.com/v1"
            slotProps={{ input: { className: 'mono' } }}
          />
          <TextField
            label={t('endpoints.apiKey')}
            type="password"
            value={form.api_key}
            onChange={(e) => setForm({ ...form, api_key: e.target.value })}
            fullWidth
            autoComplete="new-password"
            helperText={
              initial?.has_key
                ? `${t('endpoints.keySaved')} ${initial.key_hint} · ${t('endpoints.apiKeyHint')}`
                : t('endpoints.apiKeyHint')
            }
          />
          {initial?.has_key && (
            <FormControlLabel
              control={
                <Checkbox
                  checked={!!form.clear_key}
                  onChange={(e) => setForm({ ...form, clear_key: e.target.checked })}
                />
              }
              label={t('endpoints.clearKey')}
            />
          )}
          <FormControl fullWidth>
            <InputLabel>{t('endpoints.dialect')}</InputLabel>
            <Select
              label={t('endpoints.dialect')}
              value={form.dialect ?? ''}
              onChange={(e) => setForm({ ...form, dialect: e.target.value })}
            >
              <MenuItem value="">{t('endpoints.auto')}</MenuItem>
              {(dialects.data ?? []).map((d) => (
                <MenuItem key={d} value={d}>
                  {d}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <TextField
            label={t('endpoints.headers')}
            value={form.headersText}
            onChange={(e) => setForm({ ...form, headersText: e.target.value })}
            fullWidth
            multiline
            minRows={2}
            placeholder="X-Title=llm-speedtest"
            slotProps={{ input: { className: 'mono' } }}
          />
          <TextField
            label={t('endpoints.proxy')}
            value={form.proxy ?? ''}
            onChange={(e) => setForm({ ...form, proxy: e.target.value })}
            fullWidth
            placeholder="http://127.0.0.1:7890"
            slotProps={{ input: { className: 'mono' } }}
          />
          <FormControlLabel
            control={
              <Switch checked={!!form.insecure} onChange={(e) => setForm({ ...form, insecure: e.target.checked })} />
            }
            label={t('endpoints.insecure')}
          />
          {save.error && <Alert severity="error">{String(save.error)}</Alert>}
        </Stack>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button
          variant="contained"
          disableElevation
          onClick={submit}
          disabled={!form.base_url.trim() || save.isPending}
        >
          {t('common.save')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function DiscoverDialog({ result, onClose }: { result: DiscoverResult | null; onClose: () => void }) {
  const { t } = useTranslation();
  return (
    <Dialog open={!!result} onClose={onClose} fullWidth maxWidth="md">
      <DialogTitle>{t('endpoints.previewTitle')}</DialogTitle>
      <DialogContent>
        {result && (
          <Stack spacing={2}>
            <Typography variant="body2">
              {t('endpoints.discoverResult', {
                total: result.total,
                filtered: result.filtered,
                providers: arr(result.providers).length,
                field: result.provider_field,
                families: result.families,
              })}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {result.provider_rationale}
            </Typography>
            <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap>
              {arr(result.providers)
                .slice(0, 60)
                .map((p) => (
                  <Chip key={p.key} size="small" variant="outlined" label={`${p.name} · ${p.models}`} />
                ))}
            </Stack>
            <Box sx={{ maxHeight: 360, overflow: 'auto' }}>
              <Table size="small" stickyHeader>
                <TableHead>
                  <TableRow>
                    <TableCell>{t('common.model')}</TableCell>
                    <TableCell>{t('common.provider')}</TableCell>
                    <TableCell>{t('common.family')}</TableCell>
                    <TableCell>{t('run.models.familySource')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {arr(result.models)
                    .slice(0, 500)
                    .map((m) => (
                      <TableRow key={m.id} hover>
                        <TableCell className="mono">{m.id}</TableCell>
                        <TableCell>{m.provider_name}</TableCell>
                        <TableCell>{m.family_canon}</TableCell>
                        <TableCell>{m.family_source}</TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </Box>
          </Stack>
        )}
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose}>{t('common.close')}</Button>
      </DialogActions>
    </Dialog>
  );
}

function EndpointCard({
  ep,
  onEdit,
  onDelete,
  onOverrides,
}: {
  ep: Endpoint;
  onEdit: () => void;
  onDelete: () => void;
  onOverrides: () => void;
}) {
  const { t } = useTranslation();
  const check = useCheckEndpoint();
  const discover = useDiscoverEndpoint();
  const [preview, setPreview] = useState<DiscoverResult | null>(null);
  return (
    <Card>
      <CardContent sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
        <Stack direction="row" alignItems="flex-start" spacing={1}>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography variant="h6" noWrap title={ep.name}>
              {ep.name}
            </Typography>
            <Typography variant="body2" color="text.secondary" className="mono" sx={{ wordBreak: 'break-all' }}>
              {ep.base_url}
            </Typography>
          </Box>
          <Tooltip title={t('endpoints.overrides')}>
            <IconButton size="small" onClick={onOverrides}>
              <TuneRounded fontSize="small" />
            </IconButton>
          </Tooltip>
          <Tooltip title={t('common.edit')}>
            <IconButton size="small" onClick={onEdit}>
              <EditRounded fontSize="small" />
            </IconButton>
          </Tooltip>
          <Tooltip title={t('common.delete')}>
            <IconButton size="small" onClick={onDelete}>
              <DeleteOutlineRounded fontSize="small" />
            </IconButton>
          </Tooltip>
        </Stack>
        <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap>
          <Chip size="small" variant="outlined" label={ep.dialect || t('endpoints.auto')} />
          <Chip
            size="small"
            variant="outlined"
            label={ep.has_key ? `${t('endpoints.keySaved')} ${ep.key_hint}` : t('endpoints.noKey')}
          />
          {ep.insecure && <Chip size="small" color="warning" variant="outlined" label="insecure TLS" />}
          {ep.proxy && <Chip size="small" variant="outlined" label={`proxy ${ep.proxy}`} />}
        </Stack>
        {(check.isPending || discover.isPending) && <LinearProgress />}
        {check.data && <CheckResultView r={check.data} />}
        {check.error && <Alert severity="error">{String(check.error)}</Alert>}
        {discover.error && <Alert severity="error">{String(discover.error)}</Alert>}
        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
          <Button size="small" variant="outlined" onClick={() => check.mutate(ep.id)} disabled={check.isPending}>
            {check.isPending ? t('endpoints.checking') : t('endpoints.check')}
          </Button>
          <Button
            size="small"
            variant="outlined"
            onClick={() => discover.mutate({ id: ep.id }, { onSuccess: setPreview })}
            disabled={discover.isPending}
          >
            {t('endpoints.discover')}
          </Button>
          <Button
            size="small"
            variant="contained"
            disableElevation
            startIcon={<BoltRounded />}
            component={RouterLink}
            to={`/new?endpoint=${ep.id}`}
          >
            {t('endpoints.testNow')}
          </Button>
        </Stack>
      </CardContent>
      <DiscoverDialog result={preview} onClose={() => setPreview(null)} />
    </Card>
  );
}

export default function Endpoints() {
  const { t } = useTranslation();
  const endpoints = useEndpoints();
  const del = useDeleteEndpoint();
  const [dialog, setDialog] = useState<{ open: boolean; initial?: Endpoint }>({ open: false });
  const [toDelete, setToDelete] = useState<Endpoint | null>(null);
  const [overrides, setOverrides] = useState<Endpoint | null>(null);
  return (
    <>
      <PageHeader
        title={t('endpoints.title')}
        subtitle={t('endpoints.subtitle')}
        actions={
          <Button
            variant="contained"
            disableElevation
            startIcon={<AddRounded />}
            onClick={() => setDialog({ open: true })}
          >
            {t('endpoints.add')}
          </Button>
        }
      />
      {endpoints.isLoading && <LinearProgress />}
      {endpoints.isSuccess && endpoints.data.length === 0 && (
        <EmptyState
          icon={<HubRounded fontSize="inherit" />}
          title={t('endpoints.none')}
          description={t('endpoints.subtitle')}
          action={
            <Button
              variant="contained"
              disableElevation
              startIcon={<AddRounded />}
              onClick={() => setDialog({ open: true })}
            >
              {t('endpoints.add')}
            </Button>
          }
        />
      )}
      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr', xl: 'repeat(3, 1fr)' }, gap: 2 }}>
        {(endpoints.data ?? []).map((ep) => (
          <EndpointCard
            key={ep.id}
            ep={ep}
            onEdit={() => setDialog({ open: true, initial: ep })}
            onDelete={() => setToDelete(ep)}
            onOverrides={() => setOverrides(ep)}
          />
        ))}
      </Box>
      {dialog.open && (
        <EndpointDialog
          key={dialog.initial?.id ?? 'new'}
          open
          initial={dialog.initial}
          onClose={() => setDialog({ open: false })}
        />
      )}
      <ConfirmDialog
        open={!!toDelete}
        danger
        title={t('common.delete')}
        text={toDelete ? t('endpoints.deleteConfirm', { name: toDelete.name }) : ''}
        confirmLabel={t('common.delete')}
        busy={del.isPending}
        onClose={() => setToDelete(null)}
        onConfirm={() => toDelete && del.mutate(toDelete.id, { onSuccess: () => setToDelete(null) })}
      />
      <Dialog open={!!overrides} onClose={() => setOverrides(null)} fullWidth maxWidth="md">
        <DialogTitle>
          {t('endpoints.overrides')} · {overrides?.name}
        </DialogTitle>
        <DialogContent>{overrides && <OverridesPanel endpointId={overrides.id} />}</DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setOverrides(null)}>{t('common.close')}</Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
