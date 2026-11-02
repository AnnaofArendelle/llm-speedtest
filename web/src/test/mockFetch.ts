import { vi } from 'vitest';

export interface MockRoute {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  /** Exact path, path prefix ending with '*', or RegExp (matched against path + query). */
  path: string | RegExp;
  status?: number;
  /** JSON body, raw string, or a function receiving the parsed request. */
  body?: unknown | ((req: MockRequest) => unknown);
  contentType?: string;
}

export interface MockRequest {
  method: string;
  url: URL;
  path: string;
  body: unknown;
}

export interface MockFetch {
  calls: MockRequest[];
  /** Calls matching a method + path prefix. */
  find: (method: string, pathPrefix: string) => MockRequest[];
  restore: () => void;
}

function matches(route: MockRoute, method: string, path: string, full: string): boolean {
  if (route.method && route.method !== method) return false;
  if (route.path instanceof RegExp) return route.path.test(full);
  if (route.path.endsWith('*')) return path.startsWith(route.path.slice(0, -1));
  return path === route.path;
}

/** Installs a fetch stub answering from a route table (first match wins). */
export function mockFetch(routes: MockRoute[]): MockFetch {
  const calls: MockRequest[] = [];
  const impl = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
    const raw = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
    const url = new URL(raw, 'http://localhost');
    const method = (init?.method ?? 'GET').toUpperCase();
    let body: unknown = undefined;
    if (typeof init?.body === 'string') {
      try {
        body = JSON.parse(init.body);
      } catch {
        body = init.body;
      }
    }
    const req: MockRequest = { method, url, path: url.pathname, body };
    calls.push(req);
    const route = routes.find((r) => matches(r, method, url.pathname, url.pathname + url.search));
    if (!route) {
      return new Response(JSON.stringify({ error: `no mock for ${method} ${url.pathname}` }), {
        status: 404,
        headers: { 'Content-Type': 'application/json' },
      });
    }
    const payload = typeof route.body === 'function' ? (route.body as (r: MockRequest) => unknown)(req) : route.body;
    const status = route.status ?? 200;
    if (payload === undefined) return new Response('', { status });
    if (typeof payload === 'string')
      return new Response(payload, { status, headers: { 'Content-Type': route.contentType ?? 'text/plain' } });
    return new Response(JSON.stringify(payload), {
      status,
      headers: { 'Content-Type': route.contentType ?? 'application/json' },
    });
  };
  vi.stubGlobal('fetch', vi.fn(impl));
  return {
    calls,
    find: (method, prefix) => calls.filter((c) => c.method === method.toUpperCase() && c.path.startsWith(prefix)),
    restore: () => vi.unstubAllGlobals(),
  };
}
