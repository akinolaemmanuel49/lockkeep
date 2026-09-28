import { useState, useEffect } from "react";
import type { VaultItem } from "~/types/index";
import Nameplate from "./Nameplate";

interface PasswordItemModalProps {
  isOpen: boolean;
  item: VaultItem | null;
  decryptedSecret: string | null;
  onClose: () => void;
  onSave: (data: {
    name: string;
    metadata: {
      siteUrl?: string;
      identifier?: string;
      notes?: string;
    };
    secret: string;
  }) => void;
}

export default function PasswordItemModal({
  isOpen,
  item,
  decryptedSecret,
  onClose,
  onSave,
}: PasswordItemModalProps) {
  const [name, setName] = useState("");
  const [siteUrl, setSiteUrl] = useState("");
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [notes, setNotes] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");

  const isEditing = item !== null;

  useEffect(() => {
    if (isOpen && item) {
      setName(item.name);
      setSiteUrl((item.metadata?.siteUrl as string) || "");
      setIdentifier((item.metadata?.identifier as string) || "");
      setPassword(decryptedSecret || "");
      setNotes((item.metadata?.notes as string) || "");
    } else if (isOpen) {
      setName("");
      setSiteUrl("");
      setIdentifier("");
      setPassword("");
      setNotes("");
    }
    setError("");
  }, [isOpen, item, decryptedSecret]);

  const generatePassword = () => {
    const chars =
      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*";
    const length = 20;
    const array = new Uint8Array(length);
    crypto.getRandomValues(array);
    const generated = Array.from(array, (b) => chars[b % chars.length]).join("");
    setPassword(generated);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!name.trim() || !identifier.trim() || !password.trim()) {
      setError("Name, identifier, and password are required");
      return;
    }

    onSave({
      name: name.trim(),
      metadata: {
        siteUrl: siteUrl.trim() || undefined,
        identifier: identifier.trim(),
        notes: notes.trim() || undefined,
      },
      secret: password,
    });
  };

  if (!isOpen) return null;

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
            <Nameplate>{isEditing ? "Edit Record" : "New Record"}</Nameplate>
            <h2 className="mt-1.5 font-display text-xl font-medium text-ivory">
              {isEditing ? "Edit Password" : "Add Password"}
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
              className="lk-input"
              autoFocus
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
              Site URL
            </label>
            <input
              type="url"
              value={siteUrl}
              onChange={(e) => setSiteUrl(e.target.value)}
              placeholder="https://..."
              className="lk-input"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
              Identifier (Email / Username) *
            </label>
            <input
              type="text"
              value={identifier}
              onChange={(e) => setIdentifier(e.target.value)}
              placeholder="user@example.com"
              className="lk-input"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
              Password *
            </label>
            <div className="flex gap-2">
              <input
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={isEditing ? "Leave blank to keep current" : "Enter or generate password"}
                required={!isEditing}
                className="lk-input flex-1"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="lk-btn lk-btn--ghost px-3"
                aria-label={showPassword ? "Hide password" : "Show password"}
              >
                {showPassword ? (
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
              <button
                type="button"
                onClick={generatePassword}
                className="lk-btn lk-btn--brass-text px-4 text-xs"
              >
                Generate
              </button>
            </div>
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