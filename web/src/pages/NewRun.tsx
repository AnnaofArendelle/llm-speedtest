import { useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router';
import { useTranslation } from 'react-i18next';
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  Checkbox,
  Chip,
  FormControl,
  FormControlLabel,
  InputLabel,
  LinearProgress,
  MenuItem,
  Select,
  Stack,
  Switch,
  TextField,
  Typography,
} from '@mui/material';
import ExpandMoreRounded from '@mui/icons-material/ExpandMoreRounded';
import BoltRounded from '@mui/icons-material/BoltRounded';
import {
  useCheckAdhoc,
  useDialects,
  useDiscoverAdhoc,
  useDiscoverEndpoint,
  useEndpoints,
  usePresets,
  useStartRun,
} from '@/api/hooks';
import type { SuiteSpec } from '@/api/types';
import { PageHeader } from '@/components/PageHeader';
import { Section } from '@/components/Section';
import { CheckResultView } from './Endpoints';

const PROFILES = ['quick', 'standard', 'deep'] as const;
const KINDS = ['suite', 'scan', 'latency', 'capability', 'quality', 'load'] as const;
type Kind = (typeof KINDS)[number];

function splitList(s: string): string[] | null {
  const out = s
    .split(/[,\n]/)
    .map((x) => x.trim())
    .filter(Boolean);
  return out.length ? out : null;
}

function num(s: string, fallback: number): number {
  const n = Number(s);
  return s.trim() !== '' && Number.isFinite(n) ? n : fallback;
}

export default function NewRun() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const endpoints = useEndpoints();
  const presets = usePresets();
  const dialects = useDialects();
  const start = useStartRun();
  const checkAdhoc = useCheckAdhoc();
  const discoverAdhoc = useDiscoverAdhoc();
  const discoverStored = useDiscoverEndpoint();

  const [endpointId, setEndpointId] = useState(params.get('endpoint') ?? '');
  const [inline, setInline] = useState(false);
  const [inlineUrl, setInlineUrl] = useState('');
  const [inlineKey, setInlineKey] = useState('');
  const [inlineDialect, setInlineDialect] = useState('');
  const [profile, setProfile] = useState<(typeof PROFILES)[number]>('standard');
  const [kind, setKind] = useState<Kind>('suite');
  const [stages, setStages] = useState({ latency: true, capability: true, quality: true, load: false });
  const [top, setTop] = useState('');
  const [includeOwners, setIncludeOwners] = useState('');
  const [excludeOwners, setExcludeOwners] = useState('combo');
  const [idRegex, setIdRegex] = useState('');
  const [limit, setLimit] = useState('');
  const [modelsText, setModelsText] = useState('');
  const [confirmLoad, setConfirmLoad] = useState(false);
  const [lang, setLang] = useState('zh');
  const [note, setNote] = useState('');
  const [adv, setAdv] = useState({
    scanConcurrency: '',
    scanPerProvider: '',
    scanTimeout: '',
    latencySamples: '',
    latencyOutput: '',
    loadLevels: '',
    loadDuration: '',
  });

  const preset: SuiteSpec | undefined = presets.data?.[profile];
  const specs = preset?.specs as Record<string, Record<string, unknown>> | undefined;
  const ready = inline ? inlineUrl.trim() !== '' : endpointId !== '';
  const selectedEndpoint = endpoints.data?.find((e) => e.id === endpointId);

  const filter = useMemo(
    () => ({
      exclude_owners: splitList(excludeOwners),
      include_owners: splitList(includeOwners),
      id_regex: idRegex.trim(),
      exclude_regex: '',
      limit: num(limit, 0),
      models: null,
    }),
    [excludeOwners, includeOwners, idRegex, limit],
  );

  const buildSuite = (): SuiteSpec | undefined => {
    if (!preset) return undefined;
    const suite: SuiteSpec = structuredClone(preset);
    suite.profile = profile;
    suite.lang = lang;
    suite.filter = { ...suite.filter, ...filter };
    const topN = num(top, 0);
    if (kind === 'suite') {
      suite.scan = true;
      suite.latency = stages.latency;
      suite.capability = stages.capability;
      suite.quality = stages.quality;
      suite.load = stages.load;
    } else {
      suite.scan = kind === 'scan';
      suite.latency = kind === 'latency';
      suite.capability = kind === 'capability';
      suite.quality = kind === 'quality';
      suite.load = kind === 'load';
    }
    if (topN > 0) {
      suite.latency_top = topN;
      suite.capability_top = topN;
      suite.quality_top = topN;
    }
    const sp = suite.specs as Record<string, Record<string, unknown>>;
    if (adv.scanConcurrency.trim()) sp.scan.concurrency = num(adv.scanConcurrency, 10);
    if (adv.scanPerProvider.trim()) sp.scan.per_provider = num(adv.scanPerProvider, 2);
    if (adv.scanTimeout.trim()) sp.scan.timeout = num(adv.scanTimeout, 60) * 1e9;
    if (adv.latencySamples.trim()) sp.latency.samples = num(adv.latencySamples, 5);
    if (adv.latencyOutput.trim()) sp.latency.output_tokens = num(adv.latencyOutput, 128);
    if (adv.loadLevels.trim())
      sp.load.levels = (splitList(adv.loadLevels) ?? []).map((x) => num(x, 1)).filter((x) => x > 0);
    if (adv.loadDuration.trim()) sp.load.duration = num(adv.loadDuration, 30) * 1e9;
    return suite;
  };

  const submit = () => {
    const suite = buildSuite();
    if (!suite || !ready) return;
    const models = kind !== 'suite' && kind !== 'scan' ? splitList(modelsText) : null;
    const body: Record<string, unknown> = {
      kind,
      profile,
      suite,
      note,
      confirm: confirmLoad,
      models: models ?? undefined,
    };
    if (inline) {
      body.endpoint = { name: inlineUrl.trim(), base_url: inlineUrl.trim(), dialect: inlineDialect };
      body.api_key = inlineKey;
    } else {
      body.endpoint_id = endpointId;
    }
    start.mutate(body, { onSuccess: (r) => void navigate(`/runs/${r.id}`) });
  };

  const preview = () => {
    if (inline)
      discoverAdhoc.mutate({ base_url: inlineUrl.trim(), api_key: inlineKey, dialect: inlineDialect, filter });
    else discoverStored.mutate({ id: endpointId, filter });
  };
  const previewResult = inline ? discoverAdhoc.data : discoverStored.data;
  const previewPending = discoverAdhoc.isPending || discoverStored.isPending;
  const previewError = inline ? discoverAdhoc.error : discoverStored.error;
  const needsLoadConfirm = (kind === 'suite' && stages.load) || kind === 'load';
  const singleStageNeedsModels = kind !== 'suite' && kind !== 'scan';
  const d = (k: string, fallback: unknown) => String(specs?.[k.split('.')[0]]?.[k.split('.')[1]] ?? fallback);

  return (
    <>
      <PageHeader title={t('newRun.title')} subtitle={t('newRun.subtitle')} />
      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', lg: '3fr 2fr' }, gap: 2, alignItems: 'start' }}>
        <Stack spacing={2}>
          <Section title={t('newRun.endpoint')}>
            <Stack spacing={2}>
              <FormControlLabel
                control={<Switch checked={inline} onChange={(e) => setInline(e.target.checked)} />}
                label={t('newRun.inline')}
              />
              {inline ? (
                <Stack spacing={2}>
                  <TextField
                    label={t('endpoints.baseUrl')}
                    value={inlineUrl}
                    onChange={(e) => setInlineUrl(e.target.value)}
                    fullWidth
                    placeholder="http://localhost:20128/v1"
                    slotProps={{ input: { className: 'mono' } }}
                  />
                  <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
                    <TextField
                      label={t('endpoints.apiKey')}
                      type="password"
                      value={inlineKey}
                      onChange={(e) => setInlineKey(e.target.value)}
                      fullWidth
                      autoComplete="off"
                    />
                    <FormControl sx={{ minWidth: 180 }}>
                      <InputLabel>{t('endpoints.dialect')}</InputLabel>
                      <Select
                        label={t('endpoints.dialect')}
                        value={inlineDialect}
                        onChange={(e) => setInlineDialect(e.target.value)}
                      >
                        <MenuItem value="">{t('endpoints.auto')}</MenuItem>
                        {(dialects.data ?? []).map((x) => (
                          <MenuItem key={x} value={x}>
                            {x}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                  </Stack>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Button
                      size="small"
                      variant="outlined"
                      disabled={!inlineUrl.trim() || checkAdhoc.isPending}
                      onClick={() =>
                        checkAdhoc.mutate({ base_url: inlineUrl.trim(), api_key: inlineKey, dialect: inlineDialect })
                      }
                    >
                      {checkAdhoc.isPending ? t('endpoints.checking') : t('endpoints.check')}
                    </Button>
                    {checkAdhoc.data && <CheckResultView r={checkAdhoc.data} />}
                  </Stack>
                </Stack>
              ) : (
                <FormControl fullWidth>
                  <InputLabel>{t('newRun.endpoint')}</InputLabel>
                  <Select
                    label={t('newRun.endpoint')}
                    value={endpoints.data?.some((e) => e.id === endpointId) ? endpointId : ''}
                    onChange={(e) => setEndpointId(e.target.value)}
                  >
                    {(endpoints.data ?? []).map((e) => (
                      <MenuItem key={e.id} value={e.id}>
                        {e.name}{' '}
                        <span className="mono" style={{ opacity: 0.6, marginLeft: 8 }}>
                          {e.base_url}
                        </span>
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              )}
              {!inline && endpoints.isSuccess && endpoints.data.length === 0 && (
                <Alert
                  severity="info"
                  action={
                    <Button size="small" onClick={() => void navigate('/endpoints')}>
                      {t('endpoints.add')}
                    </Button>
                  }
                >
                  {t('endpoints.none')}
                </Alert>
              )}
            </Stack>
          </Section>
          <Section title={t('newRun.profile')}>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: 'repeat(3, 1fr)' }, gap: 1.5 }}>
              {PROFILES.map((p) => (
                <Card
                  key={p}
                  variant="outlined"
                  sx={{
                    boxShadow: 'none',
                    borderColor: profile === p ? 'primary.main' : 'divider',
                    borderWidth: profile === p ? 2 : 1,
                  }}
                >
                  <CardActionArea onClick={() => setProfile(p)} sx={{ p: 2, height: '100%' }}>
                    <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
                      {t(`newRun.profiles.${p}.name`)}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {t(`newRun.profiles.${p}.desc`)}
                    </Typography>
                  </CardActionArea>
                </Card>
              ))}
            </Box>
          </Section>
          <Section title={t('newRun.stages')}>
            <Stack spacing={1.5}>
              <FormControl size="small" sx={{ maxWidth: 320 }}>
                <InputLabel>{t('newRun.kind')}</InputLabel>
                <Select label={t('newRun.kind')} value={kind} onChange={(e) => setKind(e.target.value as Kind)}>
                  <MenuItem value="suite">{t('newRun.kindSuite')}</MenuItem>
                  {KINDS.filter((k) => k !== 'suite').map((k) => (
                    <MenuItem key={k} value={k}>
                      {t('newRun.kindOnly')} · {t(`kinds.${k}`)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              {kind === 'suite' && (
                <Stack>
                  <FormControlLabel control={<Switch checked disabled />} label={t('newRun.stageScan')} />
                  <FormControlLabel
                    control={
                      <Switch
                        checked={stages.latency}
                        onChange={(e) => setStages({ ...stages, latency: e.target.checked })}
                      />
                    }
                    label={t('newRun.stageLatency')}
                  />
                  <FormControlLabel
                    control={
                      <Switch
                        checked={stages.capability}
                        onChange={(e) => setStages({ ...stages, capability: e.target.checked })}
                      />
                    }
                    label={t('newRun.stageCapability')}
                  />
                  <FormControlLabel
                    control={
                      <Switch
                        checked={stages.quality}
                        onChange={(e) => setStages({ ...stages, quality: e.target.checked })}
                      />
                    }
                    label={t('newRun.stageQuality')}
                  />
                  <FormControlLabel
                    control={
                      <Switch
                        checked={stages.load}
                        onChange={(e) => setStages({ ...stages, load: e.target.checked })}
                        color="warning"
                      />
                    }
                    label={t('newRun.stageLoad')}
                  />
                  <TextField
                    size="small"
                    label={t('newRun.top')}
                    value={top}
                    onChange={(e) => setTop(e.target.value)}
                    placeholder={String(preset?.latency_top ?? 12)}
                    sx={{ maxWidth: 320, mt: 1 }}
                  />
                </Stack>
              )}
              {singleStageNeedsModels && (
                <TextField
                  label={t('newRun.modelsForStage')}
                  value={modelsText}
                  onChange={(e) => setModelsText(e.target.value)}
                  multiline
                  minRows={3}
                  fullWidth
                  slotProps={{ input: { className: 'mono' } }}
                />
              )}
              {needsLoadConfirm && (
                <FormControlLabel
                  control={<Checkbox checked={confirmLoad} onChange={(e) => setConfirmLoad(e.target.checked)} />}
                  label={t('newRun.confirmLoad')}
                />
              )}
            </Stack>
          </Section>
          {(kind === 'suite' || kind === 'scan') && (
            <Section title={t('newRun.filters')}>
              <Stack spacing={2}>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
                  <TextField
                    size="small"
                    label={t('newRun.includeOwners')}
                    value={includeOwners}
                    onChange={(e) => setIncludeOwners(e.target.value)}
                    fullWidth
                    placeholder="groq, cf"
                  />
                  <TextField
                    size="small"
                    label={t('newRun.excludeOwners')}
                    value={excludeOwners}
                    onChange={(e) => setExcludeOwners(e.target.value)}
                    fullWidth
                  />
                </Stack>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
                  <TextField
                    size="small"
                    label={t('newRun.idRegex')}
                    value={idRegex}
                    onChange={(e) => setIdRegex(e.target.value)}
                    fullWidth
                    placeholder="^(deepseek|qwen)"
                    slotProps={{ input: { className: 'mono' } }}
                  />
                  <TextField
                    size="small"
                    label={t('newRun.limit')}
                    value={limit}
                    onChange={(e) => setLimit(e.target.value)}
                    sx={{ minWidth: 200 }}
                  />
                </Stack>
                <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                  <Button size="small" variant="outlined" onClick={preview} disabled={!ready || previewPending}>
                    {t('newRun.preview')}
                  </Button>
                  {previewPending && <LinearProgress sx={{ flex: 1, minWidth: 120 }} />}
                  {previewResult && (
                    <Typography variant="body2">
                      {t('newRun.previewResult', {
                        total: previewResult.total,
                        filtered: previewResult.filtered,
                        providers: previewResult.providers.length,
                        families: previewResult.families,
                      })}
                    </Typography>
                  )}
                  {previewError && <Alert severity="error">{String(previewError)}</Alert>}
                </Stack>
                {previewResult && (
                  <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap>
                    {previewResult.providers.slice(0, 40).map((p) => (
                      <Chip key={p.key} size="small" variant="outlined" label={`${p.name} · ${p.models}`} />
                    ))}
                  </Stack>
                )}
              </Stack>
            </Section>
          )}
          <Accordion disableGutters sx={{ borderRadius: 5, '&:before': { display: 'none' }, overflow: 'hidden' }}>
            <AccordionSummary expandIcon={<ExpandMoreRounded />}>
              <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
                {t('newRun.advanced')}
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Box
                sx={{
                  display: 'grid',
                  gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' },
                  gap: 2,
                }}
              >
                <TextField
                  size="small"
                  label={t('newRun.scanConcurrency')}
                  value={adv.scanConcurrency}
                  onChange={(e) => setAdv({ ...adv, scanConcurrency: e.target.value })}
                  placeholder={d('scan.concurrency', 10)}
                />
                <TextField
                  size="small"
                  label={t('newRun.scanPerProvider')}
                  value={adv.scanPerProvider}
                  onChange={(e) => setAdv({ ...adv, scanPerProvider: e.target.value })}
                  placeholder={d('scan.per_provider', 2)}
                />
                <TextField
                  size="small"
                  label={t('newRun.scanTimeout')}
                  value={adv.scanTimeout}
                  onChange={(e) => setAdv({ ...adv, scanTimeout: e.target.value })}
                  placeholder={String(Number(d('scan.timeout', 60e9)) / 1e9)}
                />
                <TextField
                  size="small"
                  label={t('newRun.latencySamples')}
                  value={adv.latencySamples}
                  onChange={(e) => setAdv({ ...adv, latencySamples: e.target.value })}
                  placeholder={d('latency.samples', 5)}
                />
                <TextField
                  size="small"
                  label={t('newRun.latencyOutput')}
                  value={adv.latencyOutput}
                  onChange={(e) => setAdv({ ...adv, latencyOutput: e.target.value })}
                  placeholder={d('latency.output_tokens', 128)}
                />
                <TextField
                  size="small"
                  label={t('newRun.loadLevels')}
                  value={adv.loadLevels}
                  onChange={(e) => setAdv({ ...adv, loadLevels: e.target.value })}
                  placeholder={(specs?.load?.levels as number[] | undefined)?.join(', ') ?? '1, 2, 4'}
                />
                <TextField
                  size="small"
                  label={t('newRun.loadDuration')}
                  value={adv.loadDuration}
                  onChange={(e) => setAdv({ ...adv, loadDuration: e.target.value })}
                  placeholder={String(Number(d('load.duration', 30e9)) / 1e9)}
                />
                <FormControl size="small">
                  <InputLabel>{t('newRun.lang')}</InputLabel>
                  <Select label={t('newRun.lang')} value={lang} onChange={(e) => setLang(e.target.value)}>
                    <MenuItem value="zh">中文</MenuItem>
                    <MenuItem value="en">English</MenuItem>
                  </Select>
                </FormControl>
                <TextField
                  size="small"
                  label={t('newRun.note')}
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                />
              </Box>
            </AccordionDetails>
          </Accordion>
        </Stack>
        <Stack spacing={2} sx={{ position: { lg: 'sticky' }, top: { lg: 76 } }}>
          <Card>
            <CardContent sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
              <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
                {t('newRun.start')}
              </Typography>
              <Stack spacing={0.5}>
                <Typography variant="body2" color="text.secondary">
                  {t('newRun.endpoint')}: <b>{inline ? inlineUrl || '–' : (selectedEndpoint?.name ?? '–')}</b>
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {t('newRun.profile')}: <b>{t(`newRun.profiles.${profile}.name`)}</b> · {t('newRun.kind')}:{' '}
                  <b>{kind === 'suite' ? t('newRun.kindSuite') : t(`kinds.${kind}`)}</b>
                </Typography>
                {kind === 'suite' && (
                  <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap sx={{ pt: 0.5 }}>
                    <Chip size="small" label={t('kinds.scan')} color="primary" variant="outlined" />
                    {stages.latency && <Chip size="small" label={t('kinds.latency')} variant="outlined" />}
                    {stages.capability && <Chip size="small" label={t('kinds.capability')} variant="outlined" />}
                    {stages.quality && <Chip size="small" label={t('kinds.quality')} variant="outlined" />}
                    {stages.load && <Chip size="small" label={t('kinds.load')} color="warning" variant="outlined" />}
                  </Stack>
                )}
              </Stack>
              {!ready && <Alert severity="warning">{t('newRun.needEndpoint')}</Alert>}
              {start.error && <Alert severity="error">{String(start.error)}</Alert>}
              <Button
                variant="contained"
                size="large"
                disableElevation
                startIcon={<BoltRounded />}
                disabled={
                  !ready ||
                  !preset ||
                  start.isPending ||
                  (needsLoadConfirm && !confirmLoad) ||
                  (singleStageNeedsModels && !splitList(modelsText))
                }
                onClick={submit}
              >
                {start.isPending ? t('newRun.starting') : t('newRun.start')}
              </Button>
            </CardContent>
          </Card>
        </Stack>
      </Box>
    </>
  );
}
