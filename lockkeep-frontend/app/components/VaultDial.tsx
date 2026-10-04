interface VaultDialProps {
  size?: number;
  className?: string;
  spinning?: boolean;
  pointer?: string;
}

/**
 * Decorative vault-dial: an engraved combination dial, brass on coal.
 * `spinning` rotates the tick ring slowly (flywheel flavor);
 * the pointer stays at 12 o'clock.
 */
export default function VaultDial({
  size = 168,
  className = "",
  spinning = false,
  pointer = "0",
}: VaultDialProps) {
  const ticks = Array.from({ length: 96 });

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 120 120"
      fill="none"
      className={className}
      aria-hidden="true"
    >
      <defs>
        <linearGradient id="lk-dial-brass" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#f3e2ac" />
          <stop offset="45%" stopColor="#c9a227" />
          <stop offset="100%" stopColor="#6e5a15" />
        </linearGradient>
      </defs>

      {/* Flywheel ring */}
      <circle
        cx="60"
        cy="60"
        r="57"
        stroke="url(#lk-dial-brass)"
        strokeWidth="1.6"
        opacity="0.85"
      />
      <circle
        cx="60"
        cy="60"
        r="52.5"
        stroke="rgba(240,233,218,0.14)"
        strokeWidth="0.75"
      />

      <g className={spinning ? "origin-center animate-lk-spin-slow" : ""}>
        {ticks.map((_, i) => {
          const major = i % 16 === 0;
          const angle = (i / ticks.length) * Math.PI * 2;
          const r1 = major ? 43.5 : 46;
          const r2 = 52;
          const x1 = 60 + Math.sin(angle) * r1;
          const y1 = 60 - Math.cos(angle) * r1;
          const x2 = 60 + Math.sin(angle) * r2;
          const y2 = 60 - Math.cos(angle) * r2;
          return (
            <line
              key={i}
              x1={x1}
              y1={y1}
              x2={x2}
              y2={y2}
              stroke={major ? "rgba(234,210,127,0.75)" : "rgba(201,162,39,0.35)"}
              strokeWidth={major ? 1.5 : 0.8}
            />
          );
        })}
      </g>

      {/* Inner register */}
      <circle
        cx="60"
        cy="60"
        r="40"
        stroke="rgba(201,162,39,0.3)"
        strokeWidth="1.1"
      />
      <circle
        cx="60"
        cy="60"
        r="29"
        stroke="rgba(201,162,39,0.18)"
        strokeWidth="0.9"
        strokeDasharray="2 3"
      />

      {/* Fixed pointer notch at 12 o'clock */}
      <path
        d="M58 21.5h4l1.6-7h-7.2z"
        fill="#dcb955"
        opacity="0.9"
      />

      {/* Engraved center readout */}
      {pointer && (
        <text
          x="60"
          y="68"
          textAnchor="middle"
          fontFamily="var(--font-mono)"
          fontSize="22"
          fontWeight="600"
          fill="rgba(243,226,172,0.92)"
        >
          {pointer}
        </text>
      )}

      {/* Set screws */}
      <circle cx="12" cy="12" r="2" fill="rgba(201,162,39,0.5)" />
      <circle cx="108" cy="12" r="2" fill="rgba(201,162,39,0.5)" />
      <circle cx="12" cy="108" r="2" fill="rgba(201,162,39,0.5)" />
      <circle cx="108" cy="108" r="2" fill="rgba(201,162,39,0.5)" />
    </svg>
  );
}