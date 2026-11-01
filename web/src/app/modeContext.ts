import { createContext, useContext } from 'react';

export type ThemePref = 'system' | 'light' | 'dark';
export type Mode = 'light' | 'dark';

export interface ModeContextValue {
  pref: ThemePref;
  mode: Mode;
  setPref: (p: ThemePref) => void;
}

export const ModeContext = createContext<ModeContextValue>({ pref: 'system', mode: 'light', setPref: () => undefined });

export function useThemeMode() {
  return useContext(ModeContext);
}
