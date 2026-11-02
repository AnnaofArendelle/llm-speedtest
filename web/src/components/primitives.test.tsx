import { fireEvent, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { KindChip, LatencyChip, ProviderChip, ReasonChip, ScoreBar, StatusChip, TierBadge } from './badges';
import { CopyButton } from './CopyButton';
import { EmptyState } from './EmptyState';
import { KpiCard } from './KpiCard';
import { LoadingBlock } from './LoadingBlock';
import { PageHeader } from './PageHeader';
import { Section } from './Section';
import { mockFetch } from '@/test/mockFetch';
import { globalRoutes } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';

describe('primitives', () => {
  afterEach(() => vi.unstubAllGlobals());
  it('renders badges for every status and tier', async () => {
    mockFetch(globalRoutes());
    renderWithProviders(
      <>
        {['pending', 'discovering', 'running', 'done', 'cancelled', 'failed', 'interrupted'].map((s) => (
          <StatusChip key={s} status={s} />
        ))}
        {['A', 'B', 'C', undefined].map((tier, i) => (
          <TierBadge key={i} tier={tier} />
        ))}
        <LatencyChip seconds={0.5} />
        <LatencyChip seconds={2} />
        <LatencyChip seconds={7} />
        <LatencyChip seconds={20} />
        <LatencyChip seconds={45} />
        <LatencyChip seconds={0} />
        <KindChip kind="suite" />
        <ProviderChip name="groq" evidence={[{ field: 'owned_by', value: 'groq' }]} />
        <ReasonChip code="timeout" />
        <ReasonChip code="does_not_exist" showCode />
        <ScoreBar value={42.42} digits={2} suffix="%" />
      </>,
    );
    expect(screen.getByText('Running')).toBeInTheDocument();
    expect(screen.getByText('Interrupted')).toBeInTheDocument();
    expect(screen.getByText('✕')).toBeInTheDocument();
    expect(screen.getByText('500 ms')).toBeInTheDocument();
    expect(screen.getByText('45.00 s')).toBeInTheDocument();
    expect(screen.getByText('–')).toBeInTheDocument();
    expect(screen.getByText('Suite')).toBeInTheDocument();
    expect(await screen.findByText('Timeout')).toBeInTheDocument();
    expect(screen.getByText('does_not_exist')).toBeInTheDocument();
    expect(screen.getByText('42.42%')).toBeInTheDocument();
  });
  it('renders layout primitives and copies text', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true });
    renderWithProviders(
      <>
        <PageHeader title="Title" subtitle="Sub" back="/runs" actions={<button>act</button>} />
        <Section title="Sec" subtitle="sub" action={<span>a</span>}>
          body
        </Section>
        <Section noPad>plain</Section>
        <EmptyState title="Nothing" description="desc" action={<button>go</button>} />
        <EmptyState compact title="Small" />
        <KpiCard label="Models" value={2306} hint="hint" delay={0.1} />
        <KpiCard label="Text" value="1.2 s" suffix="x" />
        <KpiCard label="Custom">
          <span>child</span>
        </KpiCard>
        <LoadingBlock rows={2} />
        <CopyButton text="hello" />
      </>,
    );
    expect(screen.getByRole('heading', { name: 'Title' })).toBeInTheDocument();
    expect(screen.getByText('Nothing')).toBeInTheDocument();
    expect(screen.getByText('child')).toBeInTheDocument();
    expect(screen.getByTestId('loading-block')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Copy' }));
    expect(writeText).toHaveBeenCalledWith('hello');
    fireEvent.click(screen.getByLabelText('back'));
  });
});
