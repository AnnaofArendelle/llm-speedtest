import type { ReactNode } from 'react';
import { Card, CardContent, CardHeader } from '@mui/material';
import type { SxProps, Theme } from '@mui/material/styles';

export function Section({
  title,
  subtitle,
  action,
  children,
  sx,
  noPad,
}: {
  title?: ReactNode;
  subtitle?: ReactNode;
  action?: ReactNode;
  children: ReactNode;
  sx?: SxProps<Theme>;
  noPad?: boolean;
}) {
  return (
    <Card sx={sx}>
      {title && (
        <CardHeader
          title={title}
          subheader={subtitle}
          action={action}
          slotProps={{ title: { variant: 'subtitle1', sx: { fontWeight: 700 } }, subheader: { variant: 'body2' } }}
          sx={{ pb: 0.5, '& .MuiCardHeader-action': { alignSelf: 'center', m: 0 } }}
        />
      )}
      <CardContent sx={noPad ? { p: 0, '&:last-child': { pb: 0 } } : { pt: title ? 1.5 : 2.5 }}>{children}</CardContent>
    </Card>
  );
}
