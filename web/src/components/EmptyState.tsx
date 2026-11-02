import type { ReactNode } from 'react';
import { Box, Card, Typography } from '@mui/material';
import { Logo } from './Logo';

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
        <Box
          sx={{
            color: 'text.secondary',
            fontSize: 40,
            display: 'flex',
            mb: 0.5,
            opacity: icon ? 1 : 0.45,
            filter: icon ? 'none' : 'grayscale(0.6)',
          }}
          aria-hidden
        >
          {icon ?? <Logo size={compact ? 40 : 56} />}
        </Box>
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
