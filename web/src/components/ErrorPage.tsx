import { Box, Button, Card, CardContent, Stack, Typography } from '@mui/material';
import ErrorOutlineRounded from '@mui/icons-material/ErrorOutlineRounded';
import { isRouteErrorResponse, Link as RouterLink, useRouteError } from 'react-router';
import { useTranslation } from 'react-i18next';

/** Route-level fallback (react-router `errorElement`). */
export function ErrorPage() {
  const { t } = useTranslation();
  const error = useRouteError();
  const message = isRouteErrorResponse(error)
    ? `${error.status} ${error.statusText}`
    : error instanceof Error
      ? error.message
      : String(error ?? '');
  return (
    <Box sx={{ display: 'flex', justifyContent: 'center', py: 6 }}>
      <Card sx={{ maxWidth: 640, width: '100%' }}>
        <CardContent sx={{ p: 4, display: 'flex', flexDirection: 'column', gap: 2 }}>
          <Stack direction="row" spacing={1.5} alignItems="center">
            <ErrorOutlineRounded color="error" sx={{ fontSize: 36 }} />
            <Typography variant="h5">{t('errors.title')}</Typography>
          </Stack>
          <Typography variant="body1" color="text.secondary">
            {t('errors.description')}
          </Typography>
          {message && (
            <Box
              component="pre"
              className="mono"
              sx={{
                m: 0,
                p: 2,
                borderRadius: '12px',
                bgcolor: 'background.default',
                fontSize: 12.5,
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
              }}
            >
              {message}
            </Box>
          )}
          <Stack direction="row" spacing={1}>
            <Button variant="contained" disableElevation onClick={() => window.location.reload()}>
              {t('errors.reload')}
            </Button>
            <Button variant="outlined" component={RouterLink} to="/runs">
              {t('errors.home')}
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
}
