import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { radius } from '@/app/theme';
import {
  Box,
  Button,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from '@mui/material';
import LightModeRounded from '@mui/icons-material/LightModeRounded';
import DarkModeRounded from '@mui/icons-material/DarkModeRounded';
import BrightnessAutoRounded from '@mui/icons-material/BrightnessAutoRounded';
import { useQueryClient } from '@tanstack/react-query';
import { useEndpoints, useVersion } from '@/api/hooks';
import { useThemeMode, type ThemePref } from '@/app/modeContext';
import { PageHeader } from '@/components/PageHeader';
import { Section } from '@/components/Section';
import { OverridesPanel } from '@/components/OverridesPanel';
import { storage } from '@/lib/storage';

export default function Settings() {
  const { t, i18n } = useTranslation();
  const { pref, setPref } = useThemeMode();
  const version = useVersion();
  const endpoints = useEndpoints();
  const qc = useQueryClient();
  const [token, setToken] = useState(storage.get('token') ?? '');
  const [savedAt, setSavedAt] = useState(0);
  const [endpointId, setEndpointId] = useState('');
  const lang = i18n.language.startsWith('zh') ? 'zh' : 'en';
  const saveToken = () => {
    if (token.trim()) storage.set('token', token.trim());
    else storage.remove('token');
    setSavedAt(Date.now());
    void qc.invalidateQueries();
  };
  return (
    <>
      <PageHeader title={t('settings.title')} subtitle={t('settings.subtitle')} />
      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2 }}>
        <Section title={t('settings.theme')}>
          <ToggleButtonGroup
            exclusive
            value={pref}
            onChange={(_e, v: ThemePref | null) => v && setPref(v)}
            size="small"
          >
            <ToggleButton value="system">
              <BrightnessAutoRounded fontSize="small" sx={{ mr: 1 }} /> {t('settings.themeSystem')}
            </ToggleButton>
            <ToggleButton value="light">
              <LightModeRounded fontSize="small" sx={{ mr: 1 }} /> {t('settings.themeLight')}
            </ToggleButton>
            <ToggleButton value="dark">
              <DarkModeRounded fontSize="small" sx={{ mr: 1 }} /> {t('settings.themeDark')}
            </ToggleButton>
          </ToggleButtonGroup>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1.5 }}>
            {t('settings.reducedMotion')}: prefers-reduced-motion
          </Typography>
        </Section>
        <Section title={t('settings.language')}>
          <ToggleButtonGroup
            exclusive
            value={lang}
            size="small"
            onChange={(_e, v: string | null) => {
              if (!v) return;
              void i18n.changeLanguage(v);
              storage.set('lang', v);
              document.documentElement.lang = v === 'zh' ? 'zh-CN' : 'en';
            }}
          >
            <ToggleButton value="zh">中文</ToggleButton>
            <ToggleButton value="en">English</ToggleButton>
          </ToggleButtonGroup>
        </Section>
        <Section title={t('settings.token')} subtitle={t('settings.tokenHint')}>
          <Stack direction="row" spacing={1}>
            <TextField
              size="small"
              type="password"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              fullWidth
              autoComplete="off"
              slotProps={{ htmlInput: { 'aria-label': t('settings.token') } }}
            />
            <Button variant="contained" disableElevation onClick={saveToken}>
              {t('common.save')}
            </Button>
          </Stack>
          {savedAt > 0 && (
            <Typography variant="caption" color="success.main" sx={{ mt: 1, display: 'block' }}>
              {t('settings.saved')}
            </Typography>
          )}
        </Section>
        <Section title={t('settings.about')}>
          <Stack spacing={0.5}>
            <Typography variant="body2">
              {t('settings.version')}: <b className="mono">{version.data?.version ?? '–'}</b>{' '}
              <span className="mono" style={{ opacity: 0.7 }}>
                {version.data?.commit ? `(${version.data.commit.slice(0, 7)})` : ''}
              </span>
            </Typography>
            <Typography variant="body2" className="mono">
              {version.data?.go_version} · {version.data?.date}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              AGPL-3.0 · github.com/AnnaofArendelle/llm-speedtest
            </Typography>
          </Stack>
        </Section>
        <Section title={t('settings.skill')} subtitle={t('settings.skillHint')} sx={{ gridColumn: { md: '1 / -1' } }}>
          <Box
            component="pre"
            className="mono"
            sx={{
              m: 0,
              p: 2,
              borderRadius: radius.control,
              bgcolor: 'background.default',
              fontSize: 13,
              overflowX: 'auto',
            }}
          >
            {`llm-speedtest skill install --target claude   # Claude Code / claude.ai skills
llm-speedtest skill install --target codex    # Codex CLI
llm-speedtest skill install --target cursor   # Cursor
llm-speedtest mcp                              # MCP server over stdio`}
          </Box>
        </Section>
        <Section title={t('settings.overrides')} sx={{ gridColumn: { md: '1 / -1' } }}>
          <FormControl size="small" sx={{ minWidth: 280, mb: 2 }}>
            <InputLabel>{t('settings.selectEndpoint')}</InputLabel>
            <Select
              label={t('settings.selectEndpoint')}
              value={endpointId}
              onChange={(e) => setEndpointId(e.target.value)}
            >
              {(endpoints.data ?? []).map((e) => (
                <MenuItem key={e.id} value={e.id}>
                  {e.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          {endpointId && <OverridesPanel endpointId={endpointId} />}
        </Section>
      </Box>
    </>
  );
}
