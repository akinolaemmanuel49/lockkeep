import { useState } from "react";
import type { VaultItem } from "~/types/index";

interface VaultItemCardProps {
  item: VaultItem;
  onRevealSecret: (onReveal: (secret: string) => void) => void;
  onCopySecret: () => void;
  onEdit: (item: VaultItem) => void;
  onDelete: (id: string) => void;
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

  const metadata = item.metadata || {};
  const siteUrl = (metadata.siteUrl as string) || "";
  const identifier = (metadata.identifier as string) || "";
  const notes = (metadata.notes as string) || "";

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

  const copyIdentifier = async () => {
    if (identifier) {
      await navigator.clipboard.writeText(identifier);
    }
  };

  const maskSecret = () => "••••••••••";

  return (
    <div className="rounded-xl border border-slate-800 bg-slate-900 p-5">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="rounded bg-sky-400/10 px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-sky-400">
              {item.type.replace("_", " ")}
            </span>
          </div>
          <h3 className="mt-1 truncate text-base font-semibold text-slate-100">
            {item.name}
          </h3>
          {siteUrl && (
            <a
              href={siteUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-1 block truncate text-sm text-sky-400 hover:text-sky-300"
            >
              {siteUrl.replace(/^https?:\/\//, "")}
            </a>
          )}
        </div>
        <div className="flex gap-1">
          <button
            onClick={() => onEdit(item)}
            className="rounded-lg p-2 text-slate-500 hover:bg-slate-800 hover:text-slate-300"
            title="Edit"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
            </svg>
          </button>
          <button
            onClick={() => onDelete(item.id)}
            className="rounded-lg p-2 text-slate-500 hover:bg-slate-800 hover:text-red-400"
            title="Delete"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <polyline points="3 6 5 6 21 6" />
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
            </svg>
          </button>
        </div>
      </div>

      <div className="my-4 h-px bg-slate-800" />

      <div className="flex flex-col gap-4">
        {identifier && (
          <div>
            <span className="text-xs font-medium uppercase tracking-wider text-slate-600">
              Identifier
            </span>
            <div className="mt-1 flex items-center gap-2">
              <span className="flex-1 truncate text-sm text-slate-300">
                {identifier}
              </span>
              <button
                onClick={copyIdentifier}
                className="shrink-0 rounded p-1 text-slate-600 hover:text-slate-300"
                title="Copy identifier"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                </svg>
              </button>
            </div>
          </div>
        )}

        <div>
          <span className="text-xs font-medium uppercase tracking-wider text-slate-600">
            Password
          </span>
          <div className="mt-1 flex items-center gap-2">
            <code className="flex-1 truncate font-mono text-sm text-slate-300 tracking-wider">
              {showSecret && secret ? secret : maskSecret()}
            </code>
            <div className="flex gap-1">
              <button
                onClick={handleToggleSecret}
                className="rounded p-1 text-slate-600 hover:text-slate-300"
                title={showSecret ? "Hide" : "Show"}
              >
                {showSecret ? (
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                    <line x1="1" y1="1" x2="23" y2="23" />
                  </svg>
                ) : (
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                    <circle cx="12" cy="12" r="3" />
                  </svg>
                )}
              </button>
              <button
                onClick={handleCopySecret}
                className="rounded p-1 text-slate-600 hover:text-slate-300"
                title="Copy password"
              >
                {copied ? (
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#4ade80" strokeWidth="2">
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                ) : (
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
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
            <span className="text-xs font-medium uppercase tracking-wider text-slate-600">
              Notes
            </span>
            <p className="mt-1 text-sm text-slate-500">{notes}</p>
          </div>
        )}
      </div>

      <div className="mt-4 pt-3 border-t border-slate-800">
        <span className="text-xs text-slate-700">
          Updated {new Date(item.updatedAt).toLocaleDateString()}
        </span>
      </div>
    </div>
  );
}