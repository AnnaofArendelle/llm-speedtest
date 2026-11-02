// Helpers for fields that the Go backend may serialise as null (nil slices and
// maps). Prefer these over inline `?? []` so the intent is visible at the call
// site and the TypeScript types can stay honest (`T[] | null`).

export function arr<T>(v: readonly T[] | null | undefined): T[] {
  return v ? [...v] : [];
}

export function obj<T>(v: Record<string, T> | null | undefined): Record<string, T> {
  return v ?? {};
}

export function entries<T>(v: Record<string, T> | null | undefined): [string, T][] {
  return Object.entries(v ?? {});
}

export function keys(v: Record<string, unknown> | null | undefined): string[] {
  return Object.keys(v ?? {});
}

/** Sorted entries, largest value first. */
export function topEntries(v: Record<string, number> | null | undefined, limit = Infinity): [string, number][] {
  return entries(v)
    .sort((a, b) => b[1] - a[1])
    .slice(0, limit);
}
