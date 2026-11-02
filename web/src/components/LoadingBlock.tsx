import { Skeleton, Stack } from '@mui/material';

/** Skeleton placeholder for a card/table area while its query is loading. */
export function LoadingBlock({ height = 280, rows = 0 }: { height?: number; rows?: number }) {
  return (
    <Stack spacing={1.5} data-testid="loading-block">
      <Skeleton variant="rounded" height={height} />
      {Array.from({ length: rows }, (_, i) => (
        <Skeleton key={i} variant="rounded" height={44} />
      ))}
    </Stack>
  );
}
