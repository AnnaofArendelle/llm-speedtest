import { Suspense } from 'react';
import { NavLink, Outlet, useLocation } from 'react-router';
import { useTranslation } from 'react-i18next';
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion';
import {
  BottomNavigation,
  BottomNavigationAction,
  Box,
  Chip,
  IconButton,
  LinearProgress,
  Paper,
  Stack,
  Tooltip,
  Typography,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { alpha } from '@mui/material/styles';
import DashboardRounded from '@mui/icons-material/DashboardRounded';
import BoltRounded from '@mui/icons-material/BoltRounded';
import HubRounded from '@mui/icons-material/HubRounded';
import HistoryRounded from '@mui/icons-material/HistoryRounded';
import CompareArrowsRounded from '@mui/icons-material/CompareArrowsRounded';
import SettingsRounded from '@mui/icons-material/SettingsRounded';
import LightModeRounded from '@mui/icons-material/LightModeRounded';
import DarkModeRounded from '@mui/icons-material/DarkModeRounded';
import BrightnessAutoRounded from '@mui/icons-material/BrightnessAutoRounded';
import TranslateRounded from '@mui/icons-material/TranslateRounded';
import { useThemeMode, type ThemePref } from './modeContext';
import { useRuns } from '@/api/hooks';
import { Logo } from '@/components/Logo';
import { StatusDot } from '@/components/badges';
import { storage } from '@/lib/storage';

const RAIL_WIDTH = 92;

const NAV = [
  { key: 'dashboard', to: '/', icon: DashboardRounded, end: true },
  { key: 'newRun', to: '/new', icon: BoltRounded, end: false },
  { key: 'endpoints', to: '/endpoints', icon: HubRounded, end: false },
  { key: 'runs', to: '/runs', icon: HistoryRounded, end: false },
  { key: 'compare', to: '/compare', icon: CompareArrowsRounded, end: false },
  { key: 'settings', to: '/settings', icon: SettingsRounded, end: false },
] as const;

function isActive(path: string, to: string, end: boolean) {
  return end ? path === to : path === to || path.startsWith(to + '/');
}

function RailItem({
  to,
  label,
  icon: Icon,
  active,
}: {
  to: string;
  label: string;
  icon: typeof DashboardRounded;
  active: boolean;
}) {
  const theme = useTheme();
  const reduce = useReducedMotion();
  return (
    <Box
      component={NavLink}
      to={to}
      aria-label={label}
      aria-current={active ? 'page' : undefined}
      sx={{
        textDecoration: 'none',
        color: active ? 'primary.main' : 'text.secondary',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 0.5,
        py: 1,
        width: '100%',
        '&:hover .lst-rail-hover': { bgcolor: alpha(theme.palette.primary.main, 0.08) },
        '&:focus-visible': { outline: `2px solid ${theme.palette.primary.main}`, outlineOffset: 2, borderRadius: 2 },
      }}
    >
      <Box
        sx={{
          position: 'relative',
          width: 56,
          height: 32,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Box
          className="lst-rail-hover"
          sx={{ position: 'absolute', inset: 0, borderRadius: 999, transition: 'background-color .12s' }}
        />
        {active && (
          <motion.div
            layoutId="lst-rail-pill"
            transition={reduce ? { duration: 0 } : { type: 'spring', bounce: 0.2, duration: 0.35 }}
            style={{
              position: 'absolute',
              inset: 0,
              borderRadius: 999,
              backgroundColor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.22 : 0.14),
            }}
          />
        )}
        <Icon sx={{ position: 'relative', fontSize: 22 }} />
      </Box>
      <Typography variant="caption" sx={{ fontWeight: active ? 700 : 500, fontSize: 11.5, lineHeight: 1.1 }}>
        {label}
      </Typography>
    </Box>
  );
}

function ThemeToggle() {
  const { t } = useTranslation();
  const { pref, setPref } = useThemeMode();
  const next: Record<ThemePref, ThemePref> = { system: 'light', light: 'dark', dark: 'system' };
  const labels: Record<ThemePref, string> = {
    system: t('settings.themeSystem'),
    light: t('settings.themeLight'),
    dark: t('settings.themeDark'),
  };
  const Icon = pref === 'system' ? BrightnessAutoRounded : pref === 'light' ? LightModeRounded : DarkModeRounded;
  return (
    <Tooltip title={`${t('nav.theme')}: ${labels[pref]}`}>
      <IconButton onClick={() => setPref(next[pref])} aria-label={t('nav.theme')} size="small">
        <Icon fontSize="small" />
      </IconButton>
    </Tooltip>
  );
}

function LangToggle() {
  const { t, i18n } = useTranslation();
  const toggle = () => {
    const next = i18n.language.startsWith('zh') ? 'en' : 'zh';
    void i18n.changeLanguage(next);
    storage.set('lang', next);
    document.documentElement.lang = next === 'zh' ? 'zh-CN' : 'en';
  };
  return (
    <Tooltip title={`${t('nav.language')}: ${i18n.language.startsWith('zh') ? '中文' : 'English'}`}>
      <IconButton onClick={toggle} aria-label={t('nav.language')} size="small">
        <TranslateRounded fontSize="small" />
      </IconButton>
    </Tooltip>
  );
}

function RunningChip() {
  const { t } = useTranslation();
  const runs = useRuns(undefined, 20);
  const running = runs.data?.runs.filter((r) => r.running && !r.parent_id) ?? [];
  if (!running.length) return null;
  const first = running[0];
  return (
    <Chip
      component={NavLink}
      to={`/runs/${first.id}`}
      clickable
      size="small"
      color="primary"
      variant="outlined"
      icon={
        <Box sx={{ display: 'flex', pl: 0.75 }}>
          <StatusDot color="currentColor" pulse />
        </Box>
      }
      label={`${t('common.running')} · ${running.length}`}
      sx={{ fontWeight: 700 }}
    />
  );
}

function Fallback() {
  return <LinearProgress sx={{ position: 'fixed', top: 0, left: 0, right: 0, zIndex: 2000, height: 2 }} />;
}

export function Layout() {
  const { t } = useTranslation();
  const theme = useTheme();
  const desktop = useMediaQuery(theme.breakpoints.up('md'));
  const location = useLocation();
  const reduce = useReducedMotion();
  const current = NAV.find((n) => isActive(location.pathname, n.to, n.end)) ?? NAV[0];
  const title = location.pathname.startsWith('/runs/') ? t('run.title') : t(`nav.${current.key}`);
  return (
    <Box sx={{ minHeight: '100%', display: 'flex' }}>
      {desktop && (
        <Box
          component="nav"
          aria-label="primary"
          sx={{
            width: RAIL_WIDTH,
            position: 'fixed',
            top: 0,
            bottom: 0,
            left: 0,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            pt: 2,
            gap: 0.5,
            bgcolor: 'background.default',
            borderRight: `1px solid ${theme.palette.divider}`,
            zIndex: 10,
          }}
        >
          <Box component={NavLink} to="/" aria-label={t('app.name')} sx={{ mb: 2, display: 'flex' }}>
            <Logo size={40} />
          </Box>
          {NAV.map((n) => (
            <RailItem
              key={n.key}
              to={n.to}
              label={t(`nav.${n.key}`)}
              icon={n.icon}
              active={isActive(location.pathname, n.to, n.end)}
            />
          ))}
        </Box>
      )}
      <Box sx={{ flex: 1, minWidth: 0, ml: desktop ? `${RAIL_WIDTH}px` : 0, pb: desktop ? 4 : 10 }}>
        <Box
          component="header"
          sx={{
            position: 'sticky',
            top: 0,
            zIndex: 9,
            backdropFilter: 'blur(14px)',
            bgcolor: alpha(theme.palette.background.default, 0.78),
            borderBottom: `1px solid ${theme.palette.divider}`,
          }}
        >
          <Stack
            direction="row"
            alignItems="center"
            spacing={1}
            sx={{ height: 60, px: { xs: 2, md: 3 }, maxWidth: 1440, mx: 'auto' }}
          >
            {!desktop && <Logo size={28} />}
            <Typography variant="h6" sx={{ fontWeight: 700, flex: 1, minWidth: 0 }} noWrap>
              {title}
            </Typography>
            <RunningChip />
            <ThemeToggle />
            <LangToggle />
          </Stack>
        </Box>
        <Box component="main" sx={{ px: { xs: 2, md: 3 }, py: { xs: 2, md: 3 }, maxWidth: 1440, mx: 'auto' }}>
          <AnimatePresence mode="wait" initial={false}>
            <motion.div
              key={location.pathname}
              initial={reduce ? false : { opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={reduce ? undefined : { opacity: 0, y: -4 }}
              transition={{ duration: reduce ? 0 : 0.18, ease: 'easeOut' }}
            >
              <Suspense fallback={<Fallback />}>
                <Outlet />
              </Suspense>
            </motion.div>
          </AnimatePresence>
        </Box>
      </Box>
      {!desktop && (
        <Paper elevation={3} sx={{ position: 'fixed', bottom: 0, left: 0, right: 0, zIndex: 10, borderRadius: 0 }}>
          <BottomNavigation value={current.key} showLabels={false} sx={{ height: 64 }}>
            {NAV.map((n) => (
              <BottomNavigationAction
                key={n.key}
                value={n.key}
                component={NavLink}
                to={n.to}
                label={t(`nav.${n.key}`)}
                icon={<n.icon />}
                sx={{ minWidth: 48 }}
              />
            ))}
          </BottomNavigation>
        </Paper>
      )}
    </Box>
  );
}
