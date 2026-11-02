import { useCallback, useMemo, useState, type ReactNode } from 'react';
import { Alert, Snackbar, type AlertColor } from '@mui/material';
import { NotifyContext } from './notifyContext';

interface Notice {
  id: number;
  message: string;
  severity: AlertColor;
}

/** Global toast queue (one notice at a time, bottom-centre, M3 style). */
export function SnackbarProvider({ children }: { children: ReactNode }) {
  const [queue, setQueue] = useState<Notice[]>([]);
  const notify = useCallback((message: string, severity: AlertColor = 'info') => {
    setQueue((q) => [...q, { id: Date.now() + Math.random(), message, severity }]);
  }, []);
  const current = queue[0];
  const close = () => setQueue((q) => q.slice(1));
  const value = useMemo(() => ({ notify }), [notify]);
  return (
    <NotifyContext.Provider value={value}>
      {children}
      <Snackbar
        key={current?.id}
        open={!!current}
        autoHideDuration={4000}
        onClose={close}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={close}
          severity={current?.severity ?? 'info'}
          variant="filled"
          sx={{ borderRadius: '14px', fontWeight: 600 }}
        >
          {current?.message}
        </Alert>
      </Snackbar>
    </NotifyContext.Provider>
  );
}
