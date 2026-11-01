import { useTheme } from '@mui/material/styles';
import { tokens, type Tokens } from '@/app/theme';

/** Design tokens of the active colour mode. */
export function useTokens(): Tokens {
  const theme = useTheme();
  return tokens[theme.palette.mode];
}
