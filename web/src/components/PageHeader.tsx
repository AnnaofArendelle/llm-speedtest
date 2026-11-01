import type { ReactNode } from 'react';
import { Box, IconButton, Stack, Typography } from '@mui/material';
import ArrowBackRounded from '@mui/icons-material/ArrowBackRounded';
import { useNavigate } from 'react-router';

export function PageHeader({
  title,
  subtitle,
  actions,
  back,
}: {
  title: ReactNode;
  subtitle?: ReactNode;
  actions?: ReactNode;
  back?: string;
}) {
  const navigate = useNavigate();
  return (
    <Stack
      direction={{ xs: 'column', sm: 'row' }}
      alignItems={{ xs: 'flex-start', sm: 'center' }}
      spacing={2}
      sx={{ mb: 3 }}
    >
      {back && (
        <IconButton onClick={() => void navigate(back)} aria-label="back" sx={{ mt: { xs: 0, sm: 0.25 } }}>
          <ArrowBackRounded />
        </IconButton>
      )}
      <Box sx={{ flex: 1, minWidth: 0 }}>
        <Typography variant="h4" component="h1" sx={{ fontSize: { xs: 24, md: 30 } }}>
          {title}
        </Typography>
        {subtitle && (
          <Typography variant="body1" color="text.secondary" component="div" sx={{ mt: 0.5 }}>
            {subtitle}
          </Typography>
        )}
      </Box>
      {actions && (
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
          {actions}
        </Stack>
      )}
    </Stack>
  );
}
