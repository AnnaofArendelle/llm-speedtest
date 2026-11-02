import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ProfileView } from './ProfileCard';
import type { Profile } from '@/api/types';
import { fixture } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';

describe('ProfileView', () => {
  it('renders eight dimensions, the radar and coverage chips from a recorded profile', () => {
    const profile = fixture<Profile>('suite', 'profile-model');
    renderWithProviders(<ProfileView profile={profile} />);
    const chart = screen.getByTestId('echart');
    const option = JSON.parse(chart.getAttribute('data-option') ?? '{}') as {
      radar: { indicator: unknown[] };
      series: { data: { value: number[] }[] }[];
    };
    expect(option.radar.indicator).toHaveLength(8);
    expect(option.series[0].data[0].value).toHaveLength(8);
    expect(screen.getByText('Stability')).toBeInTheDocument();
    expect(screen.getAllByText('Not covered').length).toBeGreaterThanOrEqual(1); // the suite has no load stage
    expect(screen.getByText('Load test')).toBeInTheDocument();
  });
  it('survives null collections', () => {
    renderWithProviders(
      <ProfileView
        profile={{ id: 'x', kind: 'model', display: 'x', dimensions: null, coverage: null, measured: null, members: 1 }}
      />,
    );
    expect(screen.getByText('Capability matrix')).toBeInTheDocument();
  });
});
