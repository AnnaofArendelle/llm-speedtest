import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ErrorBoundary } from './ErrorBoundary';
import { renderWithProviders } from '@/test/render';

let shouldThrow = true;
function Bomb() {
  if (shouldThrow) throw new Error('kaboom');
  return <div>recovered</div>;
}

describe('ErrorBoundary', () => {
  it('renders an inline fallback with details and recovers on retry', () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined);
    renderWithProviders(
      <ErrorBoundary label="Overview">
        <Bomb />
      </ErrorBoundary>,
    );
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText('The “Overview” section failed')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Error details'));
    expect(screen.getByText('kaboom')).toBeInTheDocument();
    shouldThrow = false;
    fireEvent.click(screen.getByText('Retry'));
    expect(screen.getByText('recovered')).toBeInTheDocument();
    vi.restoreAllMocks();
  });
});
