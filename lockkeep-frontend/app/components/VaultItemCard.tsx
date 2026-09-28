import { useState } from "react";
import type { VaultItem } from "~/types/index";
import { getRecordType } from "~/lib/recordTypes";

interface VaultItemCardProps {
  item: VaultItem;
  onRevealSecret: (onReveal: (secret: string) => void) => void;
  onCopySecret: () => void;
  onEdit: (item: VaultItem) => void;
  onDelete: (id: string) => void;
}

const CARD_MASK = "••••••••••";
const PAN_MASK = "•••• •••• •••• ••••";

function formatPan(pan: string): string {
  const digits = pan.replace(/\D/g, "");
  return digits ? digits.match(/.{1,4}/g)?.join(" ") ?? pan : pan;
}

export default function VaultItemCard({
  item,
  onRevealSecret,
  onCopySecret,
  onEdit,
  onDelete,
}: VaultItemCardProps) {
  const [secret, setSecret] = useState<string | null>(null);
  const [showSecret, setShowSecret] = useState(false);
  const [copied, setCopied] = useState(false);

  const def = getRecordType(item.type);
  const metadata = item.metadata || {};
  const notes = (metadata.notes as string) || "";
  const siteUrl = (metadata.siteUrl as string) || "";
  const isMultiline = def.secret.multiline === true;
  const isPayment = item.type === "payment_card";
  const displayedSecret = showSecret && secret ? secret : "";

  const metadataRows = def.metadataFields
    .filter((f) => f.key !== "siteUrl")
    .map((f) => ({
      key: f.key,
      label: f.label,
      value: (metadata[f.key] as string) || "",
    }))
    .filter((r) => r.value);

  const handleToggleSecret = () => {
    if (showSecret) {
      setShowSecret(false);
      return;
    }
    onRevealSecret((s) => {
      setSecret(s);
      setShowSecret(true);
    });
  };

  const handleCopySecret = () => {
    onCopySecret();
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const copyRow = async (value: string) => {
    await navigator.clipboard.writeText(value);
  };

  const iconButton =
    "shrink-0 rounded p-1.5 text-sand-600 transition-colors hover:text-brass-300 hover:bg-brass-500/10";

  return (
    <div className="plate flex flex-col p-5">
      {/* Record header */}
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center rounded border border-brass-500/30 bg-brass-500/10 px-1.5 py-0.5 text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-brass-300">
              {def.label}
            </span>
          </div>
          <h3 className="mt-2 truncate font-display text-lg font-medium text-ivory">
            {item.name}
          </h3>
          {siteUrl && (
            <a
              href={siteUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-0.5 block truncate text-sm text-brass-400 hover:text-brass-300"
            >
              {siteUrl.replace(/^https?:\/\//, "")}
            </a>
          )}
        </div>
        <div className="flex shrink-0 gap-1">
          <button
            onClick={() => onEdit(item)}
            className={`${iconButton}`}
            title="Edit record"
          >
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
            </svg>
          </button>
          <button
            onClick={() => onDelete(item.id)}
            className={`${iconButton} hover:text-vermillion-400`}
            title="Delete record"
          >
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <polyline points="3 6 5 6 21 6" />
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
            </svg>
          </button>
        </div>
      </div>

      <div className="engraved my-4 text-xs opacity-60">
        <span className="diamond" aria-hidden="true" />
      </div>

      {/* Record body */}
      <div className="flex flex-col gap-4">
        {metadataRows.map((row) => (
          <div key={row.key}>
            <span className="text-[0.625rem] font-semibold uppercase tracking-[0.22em] text-sand-600">
              {row.label}
            </span>
            <div className="mt-1 flex items-center gap-2">
              <span className="flex-1 truncate font-mono text-sm text-sand-300">
                {row.value}
              </span>
              <button
                onClick={() => copyRow(row.value)}
                className={`${iconButton} scale-90`}
                title={`Copy ${row.label}`}
              >
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                </svg>
              </button>
            </div>
          </div>
        ))}

        <div>
          <span className="text-[0.625rem] font-semibold uppercase tracking-[0.22em] text-sand-600">
            {def.secret.label}
          </span>
          <div className="mt-1 flex items-start gap-2">
            {isMultiline ? (
              <pre className="flex-1 whitespace-pre-wrap break-words font-mono text-sm text-ivory">
                {showSecret && secret
                  ? secret
                  : `${CARD_MASK}${CARD_MASK}`}
              </pre>
            ) : (
              <code className="flex-1 truncate font-mono text-sm tracking-[0.14em] text-ivory">
                {showSecret && secret
                  ? isPayment
                    ? formatPan(secret)
                    : secret
                  : isPayment
                    ? PAN_MASK
                    : CARD_MASK}
              </code>
            )}
            <div className="flex shrink-0 gap-0.5">
              <button
                onClick={handleToggleSecret}
                className={`${iconButton} scale-90`}
                title={showSecret ? "Hide" : "Reveal"}
              >
                {showSecret ? (
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                    <line x1="1" y1="1" x2="23" y2="23" />
                  </svg>
                ) : (
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                    <circle cx="12" cy="12" r="3" />
                  </svg>
                )}
              </button>
              <button
                onClick={handleCopySecret}
                className={`${iconButton} scale-90`}
                title={`Copy ${def.secret.label}`}
              >
                {copied ? (
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#63ae85" strokeWidth="2">
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                ) : (
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                  </svg>
                )}
              </button>
            </div>
          </div>
        </div>

        {notes && (
          <div>
            <span className="text-[0.625rem] font-semibold uppercase tracking-[0.22em] text-sand-600">
              Notes
            </span>
            <p className="mt-1 text-sm text-sand-400">{notes}</p>
          </div>
        )}
      </div>

      <div className="mt-4 flex items-center gap-2 border-t border-brass-500/10 pt-3">
        <span className="text-[0.625rem] uppercase tracking-[0.18em] text-sand-600">
          Index Record
        </span>
        <span className="h-px flex-1 bg-brass-500/15" />
        <span className="font-mono text-[0.625rem] text-sand-600">
          {new Date(item.updatedAt).toLocaleDateString()}
        </span>
      </div>
    </div>
  );
}