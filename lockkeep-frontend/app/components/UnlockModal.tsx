import { useState } from "react";
import { useVault } from "~/providers/vault";
import { useToast } from "~/providers/toast";

interface UnlockModalProps {
  isOpen: boolean;
  onUnlock: () => void;
  onCancel: () => void;
}

export default function UnlockModal({
  isOpen,
  onUnlock,
  onCancel,
}: UnlockModalProps) {
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const { unlockVault } = useVault();
  const { addToast } = useToast();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      await unlockVault(password);
      addToast("Vault unlocked", "success");
      setPassword("");
      onUnlock();
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to unlock vault",
        "error",
      );
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4"
      onClick={onCancel}
    >
      <div
        className="w-full max-w-md rounded-2xl border border-slate-700 bg-slate-900 p-8"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-8 text-center">
          <div className="mb-4 inline-flex rounded-full bg-amber-500/10 p-3 text-amber-400">
            <svg
              width="32"
              height="32"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
              <path d="M7 11V7a5 5 0 0 1 10 0v4" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-slate-100">Vault Locked</h1>
          <p className="mt-2 text-sm text-slate-500">
            Enter your vault password to continue.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          <div className="flex flex-col gap-2">
            <label className="text-sm font-medium text-slate-400">
              Vault Password
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Your vault password"
              required
              autoFocus
              className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-3 text-sm text-slate-200 placeholder:text-slate-600 focus:border-sky-400 focus:outline-none"
            />
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="rounded-lg bg-amber-400 py-3 text-sm font-semibold text-slate-950 hover:bg-amber-300 disabled:opacity-50 cursor-pointer"
          >
            {isLoading ? "Unlocking..." : "Unlock Vault"}
          </button>

          <button
            type="button"
            onClick={onCancel}
            className="rounded-lg border border-slate-700 py-3 text-sm font-medium text-slate-400 hover:text-slate-200 cursor-pointer"
          >
            Cancel
          </button>
        </form>
      </div>
    </div>
  );
}
