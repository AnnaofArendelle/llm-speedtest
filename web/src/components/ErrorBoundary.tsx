import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Alert, AlertTitle, Box, Button, Collapse, Typography } from '@mui/material';
import { withTranslation, type WithTranslation } from 'react-i18next';

interface Props extends WithTranslation {
  children: ReactNode;
  /** Short description of the region, shown in the fallback title. */
  label?: string;
  /** When any of these values change the boundary resets itself. */
  resetKeys?: unknown[];
  onError?: (error: Error, info: ErrorInfo) => void;
}

interface State {
  error: Error | null;
  showDetails: boolean;
}

/**
 * Local error boundary: a crash inside one tab or drawer renders an inline
 * fallback instead of replacing the whole page with the router error screen.
 */
class ErrorBoundaryImpl extends Component<Props, State> {
  state: State = { error: null, showDetails: false };

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    this.props.onError?.(error, info);
    if (import.meta.env.DEV) console.error('[ErrorBoundary]', error, info.componentStack);
  }

  componentDidUpdate(prev: Props) {
    if (
      this.state.error &&
      prev.resetKeys !== this.props.resetKeys &&
      !shallowEqual(prev.resetKeys, this.props.resetKeys)
    ) {
      this.setState({ error: null, showDetails: false });
    }
  }

  render() {
    const { error, showDetails } = this.state;
    const { t, label, children } = this.props;
    if (!error) return children;
    return (
      <Alert
        severity="error"
        variant="outlined"
        role="alert"
        sx={{ borderRadius: '20px', alignItems: 'flex-start' }}
        action={
          <Button color="inherit" size="small" onClick={() => this.setState({ error: null, showDetails: false })}>
            {t('errors.retry')}
          </Button>
        }
      >
        <AlertTitle>{label ? t('errors.titleIn', { label }) : t('errors.title')}</AlertTitle>
        <Typography variant="body2">{t('errors.description')}</Typography>
        <Button
          size="small"
          color="inherit"
          sx={{ mt: 1, px: 0 }}
          onClick={() => this.setState({ showDetails: !showDetails })}
        >
          {t('errors.details')}
        </Button>
        <Collapse in={showDetails}>
          <Box
            component="pre"
            className="mono"
            sx={{ m: 0, mt: 1, fontSize: 12, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}
          >
            {error.message}
          </Box>
        </Collapse>
      </Alert>
    );
  }
}

function shallowEqual(a?: unknown[], b?: unknown[]) {
  if (a === b) return true;
  if (!a || !b || a.length !== b.length) return false;
  return a.every((v, i) => Object.is(v, b[i]));
}

export const ErrorBoundary = withTranslation()(ErrorBoundaryImpl);
