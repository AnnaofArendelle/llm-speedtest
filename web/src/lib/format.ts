export function fmtSeconds(s: number | undefined | null, digits = 2): string {
  if (s === undefined || s === null || s <= 0 || Number.isNaN(s)) return '–';
  if (s < 1) return `${Math.round(s * 1000)} ms`;
  return `${s.toFixed(digits)} s`;
}

export function fmtDurationS(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) return '–';
  const s = Math.round(seconds);
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const r = s % 60;
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m ${r}s`;
  return `${r}s`;
}

export function fmtPct(num: number, den: number, digits = 1): string {
  if (!den) return '–';
  return `${((100 * num) / den).toFixed(digits)}%`;
}

export function fmtRate(rate: number, digits = 0): string {
  if (!Number.isFinite(rate)) return '–';
  return `${(100 * rate).toFixed(digits)}%`;
}

export function fmtNumber(n: number | undefined | null, digits = 1): string {
  if (n === undefined || n === null || !Number.isFinite(n)) return '–';
  return n.toFixed(digits);
}

export function fmtContext(n?: number): string {
  if (!n || n <= 0) return '–';
  if (n >= 1000) return `${Math.round(n / 1000)}k`;
  return String(n);
}

export function fmtDate(iso?: string): string {
  if (!iso) return '–';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime()) || d.getFullYear() < 2000) return '–';
  return d.toLocaleString();
}

export function fmtRelative(iso?: string, now = Date.now()): string {
  if (!iso) return '–';
  const d = new Date(iso).getTime();
  if (Number.isNaN(d) || d < 946684800000) return '–';
  const diff = Math.max(0, now - d) / 1000;
  if (diff < 60) return `${Math.round(diff)}s`;
  if (diff < 3600) return `${Math.round(diff / 60)}m`;
  if (diff < 86400) return `${Math.round(diff / 3600)}h`;
  return `${Math.round(diff / 86400)}d`;
}

export function truncate(s: string | undefined, n: number): string {
  if (!s) return '';
  return s.length <= n ? s : s.slice(0, n - 1) + '…';
}

export function durationBetween(a?: string, b?: string): string {
  if (!a || !b) return '–';
  const s = (new Date(b).getTime() - new Date(a).getTime()) / 1000;
  return fmtDurationS(s);
}
