type Listener = (e: MessageEvent) => void;

/** Minimal EventSource stand-in; tests drive it with open()/emit()/fail(). */
export class FakeEventSource {
  static instances: FakeEventSource[] = [];
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSED = 2;
  readonly CONNECTING = 0;
  readonly OPEN = 1;
  readonly CLOSED = 2;
  readonly url: string;
  readyState = 0;
  withCredentials = false;
  onopen: ((e: Event) => void) | null = null;
  onerror: ((e: Event) => void) | null = null;
  onmessage: ((e: MessageEvent) => void) | null = null;
  private listeners = new Map<string, Set<Listener>>();

  constructor(url: string | URL) {
    this.url = String(url);
    FakeEventSource.instances.push(this);
  }
  static reset() {
    FakeEventSource.instances = [];
  }
  static latest(): FakeEventSource {
    const es = FakeEventSource.instances[FakeEventSource.instances.length - 1];
    if (!es) throw new Error('no EventSource was created');
    return es;
  }
  addEventListener(type: string, fn: Listener) {
    if (!this.listeners.has(type)) this.listeners.set(type, new Set());
    this.listeners.get(type)!.add(fn);
  }
  removeEventListener(type: string, fn: Listener) {
    this.listeners.get(type)?.delete(fn);
  }
  dispatchEvent(): boolean {
    return true;
  }
  close() {
    this.readyState = FakeEventSource.CLOSED;
  }
  // ---- test helpers
  open() {
    this.readyState = FakeEventSource.OPEN;
    this.onopen?.(new Event('open'));
  }
  emit(type: string, data: unknown) {
    const ev = new MessageEvent(type, { data: JSON.stringify(data) });
    this.listeners.get(type)?.forEach((fn) => fn(ev));
  }
  fail() {
    this.readyState = FakeEventSource.CLOSED;
    const ev = new MessageEvent('error');
    this.listeners.get('error')?.forEach((fn) => fn(ev));
    this.onerror?.(ev);
  }
}
