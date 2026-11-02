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

// framer-motion drives transforms through the Web Animations API when Element.animate exists;
// happy-dom implements it, and cancelling those animations on unmount rejects their `finished`
// promises with AbortError (reported by vitest as unhandled errors). Force the JS fallback.
Object.defineProperty(Element.prototype, 'animate', { value: undefined, configurable: true, writable: true });

// ECharts needs a real canvas; render a placeholder that exposes the option for assertions.
vi.mock('@/charts/EChart', () => ({
  EChart: ({ option, height }: { option: unknown; height?: number | string }) => (
    <div data-testid="echart" data-option={JSON.stringify(option)} style={{ height }} />
  ),
}));
