import { useCallback, useEffect, useMemo, useState, useSyncExternalStore, type ReactNode } from 'react';
import { storage } from '@/lib/storage';
import { ModeContext, type Mode, type ThemePref } from './modeContext';

function query() {
  return window.matchMedia('(prefers-color-scheme: dark)');
}
function subscribe(cb: () => void) {
  const mq = query();
  mq.addEventListener('change', cb);
  return () => mq.removeEventListener('change', cb);
}
function getSystem(): Mode {
  return query().matches ? 'dark' : 'light';
}
function getServer(): Mode {
  return 'light';
}

export function ThemeModeProvider({ children }: { children: ReactNode }) {
  const [pref, setPrefState] = useState<ThemePref>(() => {
    const s = storage.get('theme');
    return s === 'light' || s === 'dark' ? s : 'system';
  });
  const system = useSyncExternalStore(subscribe, getSystem, getServer);
  const mode: Mode = pref === 'system' ? system : pref;
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', mode);
    document.documentElement.style.backgroundColor = mode === 'dark' ? '#0B0F19' : '#FBFCFE';
  }, [mode]);
  const setPref = useCallback((p: ThemePref) => {
    setPrefState(p);
    storage.set('theme', p);
  }, []);
  const value = useMemo(() => ({ pref, mode, setPref }), [pref, mode, setPref]);
  return <ModeContext.Provider value={value}>{children}</ModeContext.Provider>;
}
