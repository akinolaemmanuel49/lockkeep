interface NameplateProps {
  children: React.ReactNode;
  tone?: "brass" | "muted" | "danger" | "patina";
  className?: string;
}

const tones = {
  brass: "text-brass-400",
  muted: "text-sand-500",
  danger: "text-vermillion-400",
  patina: "text-patina-400",
};

/**
 * Engraved eyebrow label: "── LABEL" in spaced small caps.
 */
export default function Nameplate({
  children,
  tone = "brass",
  className = "",
}: NameplateProps) {
  return (
    <span
      className={`inline-flex items-center gap-2 text-[0.6875rem] font-semibold uppercase tracking-[0.3em] ${tones[tone]} ${className}`}
    >
      <span className="h-px w-6 bg-current opacity-60" aria-hidden="true" />
      {children}
    </span>
  );
}