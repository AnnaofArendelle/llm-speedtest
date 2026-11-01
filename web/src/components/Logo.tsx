export function Logo({ size = 36 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 64 64" role="img" aria-label="LLM Speedtest">
      <defs>
        <linearGradient id="lst-logo-g" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0" stopColor="#22D3EE" />
          <stop offset="1" stopColor="#2563EB" />
        </linearGradient>
      </defs>
      <rect x="4" y="4" width="56" height="56" rx="16" fill="url(#lst-logo-g)" />
      <path
        d="M15 42a17 17 0 0 1 34 0"
        fill="none"
        stroke="#fff"
        strokeOpacity="0.92"
        strokeWidth="5"
        strokeLinecap="round"
      />
      <path d="M32 42L43 27" stroke="#fff" strokeWidth="4.5" strokeLinecap="round" />
      <circle cx="32" cy="42" r="4.2" fill="#fff" />
    </svg>
  );
}
