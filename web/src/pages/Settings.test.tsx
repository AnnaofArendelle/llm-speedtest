import { fireEvent, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Settings from './Settings';
import { mockFetch } from '@/test/mockFetch';
import { globalRoutes } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';
import { storage } from '@/lib/storage';
import i18n from '@/app/i18n';

describe('Settings', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    storage.remove('token');
    storage.remove('theme');
    void i18n.changeLanguage('en');
  });
  it('saves the access token and switches theme and language', async () => {
    mockFetch(globalRoutes());
    renderWithProviders(<Settings />);
    expect(await screen.findByText('Settings')).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('API access token', { exact: false }) ?? screen.getByDisplayValue(''), {
      target: { value: 'secret' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    expect(storage.get('token')).toBe('secret');
    fireEvent.click(screen.getByRole('button', { name: /Dark/ }));
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
    expect(storage.get('theme')).toBe('dark');
    fireEvent.click(screen.getByRole('button', { name: '中文' }));
    expect(await screen.findByText('设置')).toBeInTheDocument();
  });
});
