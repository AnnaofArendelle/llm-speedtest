import { storage } from '@/lib/storage';

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

function headers(json = false): HeadersInit {
  const h: Record<string, string> = {};
  if (json) h['Content-Type'] = 'application/json';
  const token = storage.get('token');
  if (token) h['Authorization'] = `Bearer ${token}`;
  return h;
}

export async function api<T>(path: string, init?: RequestInit & { json?: unknown }): Promise<T> {
  const { json, ...rest } = init ?? {};
  const res = await fetch(path, {
    ...rest,
    headers: { ...headers(json !== undefined), ...(rest.headers ?? {}) },
    body: json !== undefined ? JSON.stringify(json) : rest.body,
  });
  const text = await res.text();
  if (!res.ok) {
    let msg = text || res.statusText;
    try {
      const parsed = JSON.parse(text) as { error?: string };
      if (parsed.error) msg = parsed.error;
    } catch {
      /* raw text */
    }
    throw new ApiError(res.status, msg);
  }
  if (!text) return undefined as T;
  return JSON.parse(text) as T;
}

export async function apiText(path: string): Promise<string> {
  const res = await fetch(path, { headers: headers() });
  if (!res.ok) throw new ApiError(res.status, await res.text());
  return res.text();
}

export function withToken(path: string): string {
  const token = storage.get('token');
  if (!token) return path;
  return path + (path.includes('?') ? '&' : '?') + 'token=' + encodeURIComponent(token);
}

/** Encodes a path value that may itself contain slashes (model ids), keeping the slashes. */
export function encodePath(value: string): string {
  return value.split('/').map(encodeURIComponent).join('/');
}
