import { createContext, useContext } from 'react';
import type { AlertColor } from '@mui/material';

export interface NotifyContextValue {
  notify: (message: string, severity?: AlertColor) => void;
}

export const NotifyContext = createContext<NotifyContextValue>({ notify: () => undefined });

/** Shows a toast through the nearest SnackbarProvider. */
export function useNotify() {
  return useContext(NotifyContext);
}
