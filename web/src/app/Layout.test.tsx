import { fireEvent, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { Layout } from './Layout';
import i18n from './i18n';
import { mockFetch } from '@/test/mockFetch';
import { globalRoutes } from '@/test/fixtures';
import { renderWithProviders } from '@/test/render';
import { storage } from '@/lib/storage';

describe('Layout', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    storage.remove('theme');
    storage.remove('lang');
    void i18n.changeLanguage('en');
  });
  it('renders navigation, cycles the theme and toggles the language', async () => {
    mockFetch(globalRoutes());
    renderWithProviders(<Layout />, { route: '/runs', path: '*' });
    expect(await screen.findByRole('heading', { name: 'Runs' })).toBeInTheDocument();
    const hrefs = screen.getAllByRole('link').map((a) => a.getAttribute('href'));
    for (const to of ['/', '/new', '/endpoints', '/runs', '/compare', '/settings']) expect(hrefs).toContain(to);
    const theme = screen.getByRole('button', { name: 'Theme' });
    fireEvent.click(theme); // system -> light
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
    fireEvent.click(theme); // light -> dark
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
    fireEvent.click(screen.getByRole('button', { name: 'Language' }));
    expect(await screen.findByRole('heading', { name: '运行记录' })).toBeInTheDocument();
    expect(document.documentElement.lang).toBe('zh-CN');
  });
});
