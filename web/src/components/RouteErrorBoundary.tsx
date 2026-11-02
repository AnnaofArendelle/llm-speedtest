import { useRouteError } from 'react-router';
import { Box, Button } from '@mui/material';
import ErrorOutlineRounded from '@mui/icons-material/ErrorOutlineRounded';
import RefreshRounded from '@mui/icons-material/RefreshRounded';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@/components/EmptyState';

export function RouteErrorBoundary() {
  const error = useRouteError();
  const { t } = useTranslation();
  const message = error instanceof Error ? error.message : String(error);

  return (
    <Box sx={{ p: 4 }}>
      <EmptyState
        icon={<ErrorOutlineRounded fontSize="inherit" color="error" />}
        title={t('common.error', { defaultValue: 'Something went wrong' })}
        description={message}
        action={
          <Button variant="outlined" startIcon={<RefreshRounded />} onClick={() => window.location.reload()}>
            {t('common.retry', { defaultValue: 'Reload' })}
          </Button>
        }
      />
    </Box>
  );
}
