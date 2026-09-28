interface BrandLockProps {
  className?: string;
}

/**
 * The LockKeep mark: a vault-door plate with a keyhole,
 * engraved in brass.
 */
export default function BrandLock({ className = "" }: BrandLockProps) {
  return (
    <svg
      width="32"
      height="32"
      viewBox="0 0 32 32"
      fill="none"
      className={className}
      aria-hidden="true"
    >
      <defs>
        <linearGradient id="lk-brand-brass" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#f3e2ac" />
          <stop offset="50%" stopColor="#c9a227" />
          <stop offset="100%" stopColor="#6e5a15" />
        </linearGradient>
      </defs>

      {/* Door plate */}
      <circle
        cx="16"
        cy="16"
        r="14.5"
        stroke="url(#lk-brand-brass)"
        strokeWidth="1.6"
        fill="rgba(25,21,15,0.9)"
      />
      <circle
        cx="16"
        cy="16"
        r="11"
        stroke="rgba(240,233,218,0.12)"
        strokeWidth="0.8"
      />

      {/* Keyhole */}
      <g fill="rgba(220,185,85,0.95)">
        <circle cx="16" cy="14.5" r="3.6" />
        <path d="M13.9 17.5h4.2l1.6 4.4h-7.4z" />
      </g>

      {/* Set screws */}
      <circle cx="8" cy="8" r="1.1" fill="rgba(201,162,39,0.55)" />
      <circle cx="24" cy="8" r="1.1" fill="rgba(201,162,39,0.55)" />
      <circle cx="8" cy="24" r="1.1" fill="rgba(201,162,39,0.55)" />
      <circle cx="24" cy="24" r="1.1" fill="rgba(201,162,39,0.55)" />
    </svg>
  );
}