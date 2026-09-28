import { strengthColors, strengthLabels } from "~/utils/calculatePasswordStrength";

export default function PasswordStrength({
  password,
  score,
}: {
  password: string;
  score: number;
}) {
  if (!password) return null;

  return (
    <div className="flex items-center gap-3">
      <div className="flex h-1.5 flex-1 gap-1 overflow-hidden">
        {Array.from({ length: 5 }).map((_, i) => (
          <span
            key={i}
            className={`h-full flex-1 rounded-full transition-colors duration-300 ${
              i < score ? strengthColors[score] : "bg-sand-600/40"
            }`}
          />
        ))}
      </div>
      <span className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-400">
        {strengthLabels[score]}
      </span>
    </div>
  );
}