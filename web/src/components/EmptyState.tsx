import type { ReactNode } from 'react';
import { Box, Card, Typography } from '@mui/material';

export function EmptyState({
  icon,
  title,
  description,
  action,
  compact,
}: {
  icon?: ReactNode;
  title: ReactNode;
  description?: ReactNode;
  action?: ReactNode;
  compact?: boolean;
}) {
  return (
    <Card variant="outlined" sx={{ borderStyle: 'dashed', bgcolor: 'transparent', boxShadow: 'none' }}>
      <Box
        sx={{
          p: compact ? 3 : 6,
          textAlign: 'center',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 1,
        }}
      >
        {icon && <Box sx={{ color: 'text.secondary', fontSize: 40, display: 'flex', mb: 0.5 }}>{icon}</Box>}
        <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
          {title}
        </Typography>
        {description && (
          <Typography variant="body2" color="text.secondary" sx={{ maxWidth: 520 }}>
            {description}
          </Typography>
        )}
        {action && <Box sx={{ mt: 1.5 }}>{action}</Box>}
      </Box>
    </Card>
  );
}
