import { useEffect, useState } from "react";
import type { VaultItem, VaultItemInput, VaultItemType } from "~/types/index";
import { getRecordType } from "~/lib/recordTypes";
import Nameplate from "./Nameplate";

interface ItemModalProps {
  isOpen: boolean;
  type: VaultItemType;
  item: VaultItem | null;
  decryptedSecret: string | null;
  onClose: () => void;
  onSave: (data: VaultItemInput) => void;
}

export default function ItemModal({
  isOpen,
  type,
  item,
  decryptedSecret,
  onClose,
  onSave,
}: ItemModalProps) {
  const def = getRecordType(type);
  const isEditing = item !== null;

  const [name, setName] = useState("");
  const [metadataValues, setMetadataValues] = useState<Record<string, string>>({});
  const [secret, setSecret] = useState("");
  const [notes, setNotes] = useState("");
  const [showSecret, setShowSecret] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (isOpen) {
      const meta: Record<string, string> = {};
      for (const field of def.metadataFields) {
        meta[field.key] = ((item?.metadata?.[field.key] as string) ||
          "") as string;
      }
      setName(item?.name ?? "");
      setMetadataValues(meta);
      setSecret(decryptedSecret ?? "");
      setNotes(((item?.metadata?.notes as string) || "") as string);
      setShowSecret(false);
      setError("");
    }
  }, [isOpen, item, decryptedSecret, type]);

  if (!isOpen) return null;

  const generatePassword = () => {
    const chars =
      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*";
    const length = 20;
    const array = new Uint8Array(length);
    crypto.getRandomValues(array);
    setSecret(Array.from(array, (b) => chars[b % chars.length]).join(""));
  };

  const setMetadataValue = (key: string, value: string) => {
    setMetadataValues((prev) => ({ ...prev, [key]: value }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!name.trim()) {
      setError("Record name is required");
      return;
    }
    for (const field of def.metadataFields) {
      if (field.required && !(metadataValues[field.key] || "").trim()) {
        setError(`${field.label} is required`);
        return;
      }
    }
    if (!isEditing && !secret.trim()) {
      setError(`${def.secret.label} is required`);
      return;
    }

    const metadata: Record<string, unknown> = {};
    for (const field of def.metadataFields) {
      const value = (metadataValues[field.key] || "").trim();
      if (value) metadata[field.key] = value;
    }
    if (notes.trim()) metadata.notes = notes.trim();

    onSave({
      type,
      name: name.trim(),
      metadata,
      secret: secret.trim(),
    });
  };

  const inputClass = "lk-input";
  const footerButtonClass = "lk-btn";

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-coal-950/85 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="plate w-full max-w-lg p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center justify-between border-b border-brass-500/15 pb-4">
          <div>
            <Nameplate>{def.label}</Nameplate>
            <h2 className="mt-1.5 font-display text-xl font-medium text-ivory">
              {isEditing ? "Edit Record" : "New Record"}
            </h2>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-sand-600 transition-colors hover:text-ivory"
            aria-label="Close"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-lg border border-vermillion-500/30 bg-vermillion-500/10 px-4 py-2 text-sm text-vermillion-300">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
              Record Name *
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., GitHub, AWS, Stripe"
              className={inputClass}
              autoFocus
            />
          </div>

          {def.metadataFields.map((field) => (
            <div key={field.key} className="flex flex-col gap-1.5">
              <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
                {field.label}
                {field.required && " *"}
              </label>
              <input
                type={field.kind === "url" ? "url" : field.kind === "month" ? "month" : "text"}
                value={metadataValues[field.key] || ""}
                onChange={(e) => setMetadataValue(field.key, e.target.value)}
                placeholder={field.placeholder}
                className={inputClass}
              />
            </div>
          ))}

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
              {def.secret.label}
              {!isEditing && " *"}
            </label>
            {def.secret.multiline ? (
              <textarea
                value={secret}
                onChange={(e) => setSecret(e.target.value)}
                placeholder={def.secret.placeholder}
                rows={5}
                className={`${inputClass} resize-y font-mono text-xs leading-relaxed`}
              />
            ) : (
              <div className="flex gap-2">
                <input
                  type={showSecret ? "text" : "password"}
                  value={secret}
                  onChange={(e) => setSecret(e.target.value)}
                  placeholder={
                    isEditing
                      ? "Leave blank to keep current"
                      : def.secret.placeholder
                  }
                  className={`${inputClass} flex-1`}
                />
                <button
                  type="button"
                  onClick={() => setShowSecret(!showSecret)}
                  className={footerButtonClass}
                  aria-label={showSecret ? "Hide" : "Show"}
                >
                  {showSecret ? (
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                      <line x1="1" y1="1" x2="23" y2="23" />
                    </svg>
                  ) : (
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                      <circle cx="12" cy="12" r="3" />
                    </svg>
                  )}
                </button>
              </div>
            )}
            {def.secret.generate && (
              <button
                type="button"
                onClick={generatePassword}
                className="lk-btn lk-btn--brass-text self-start px-4 text-xs"
              >
                Generate
              </button>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
              Notes
            </label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Additional information..."
              rows={3}
              className="lk-input resize-y min-h-[76px]"
            />
          </div>

          <div className="mt-2 flex justify-end gap-3 border-t border-brass-500/15 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="lk-btn lk-btn--ghost"
            >
              Cancel
            </button>
            <button type="submit" className="lk-btn lk-btn--primary">
              {isEditing ? "Update Record" : "File Record"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}