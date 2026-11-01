import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ThemeProvider } from '@mui/material/styles';
import '@/app/i18n';
import { buildTheme } from '@/app/theme';
import { LatencyChip, ScoreBar, StatusChip, TierBadge } from './badges';

function wrap(ui: React.ReactElement) {
  return render(<ThemeProvider theme={buildTheme('dark')}>{ui}</ThemeProvider>);
}

describe('badges', () => {
  it('renders tier badges with a fallback for unavailable models', () => {
    wrap(
      <>
        <TierBadge tier="A" />
        <TierBadge />
      </>,
    );
    expect(screen.getByText('A')).toBeInTheDocument();
    expect(screen.getByText('✕')).toBeInTheDocument();
  });
  it('formats latency chips', () => {
    wrap(<LatencyChip seconds={0.42} />);
    expect(screen.getByText('420 ms')).toBeInTheDocument();
  });
  it('renders run status and score bars', () => {
    wrap(
      <>
        <StatusChip status="running" />
        <ScoreBar value={79.7} />
      </>,
    );
    expect(screen.getByText('79.7')).toBeInTheDocument();
    expect(screen.getByText(/running|运行中/i)).toBeInTheDocument();
  });
});
