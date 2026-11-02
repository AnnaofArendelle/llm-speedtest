import '@testing-library/jest-dom/vitest';
import { vi } from 'vitest';
import i18n from '@/app/i18n';
import { FakeEventSource } from './fakeEventSource';

// Deterministic language for assertions (the i18n test checks both locales).
void i18n.changeLanguage('en');

// happy-dom lacks ResizeObserver (MUI DataGrid) and EventSource (SSE).
class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}
Object.assign(globalThis, {
  ResizeObserver: globalThis.ResizeObserver ?? ResizeObserverStub,
  EventSource: FakeEventSource,
});

// ECharts needs a real canvas; render a placeholder that exposes the option for assertions.
vi.mock('@/charts/EChart', () => ({
  EChart: ({ option, height }: { option: unknown; height?: number | string }) => (
    <div data-testid="echart" data-option={JSON.stringify(option)} style={{ height }} />
  ),
}));
