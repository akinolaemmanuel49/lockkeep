import { useState } from "react";
import { Link } from "react-router";
import { useVault } from "~/providers/vault";
import { useToast } from "~/providers/toast";
import { useAuth } from "~/providers/auth";
import VaultDial from "~/components/VaultDial";
import Nameplate from "~/components/Nameplate";

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
  const { user } = useAuth();

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
      className="fixed inset-0 z-50 flex items-center justify-center bg-coal-950/85 p-4 backdrop-blur-sm"
      onClick={onCancel}
    >
      <div
        className="plate w-full max-w-md p-8"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-8 text-center">
          <VaultDial size={104} className="mx-auto text-brass-500" />
          <Nameplate className="mt-5 justify-center">
            Combination Required
          </Nameplate>
          <h1 className="mt-3 font-display text-2xl font-semibold tracking-tight text-ivory">
            Vault Locked
          </h1>
          <p className="mt-2 text-sm text-sand-500">
            Enter your vault password to continue.
          </p>
        </div>

        {user && !user.hasMasterPassword ? (
          <div className="flex flex-col gap-5 text-center">
            <p className="text-sm leading-relaxed text-sand-500">
              Your vault hasn't been sealed yet. Set a master password before
              you can start storing secrets.
            </p>
            <Link to="/setup" className="lk-btn lk-btn--primary w-full py-3">
              Set Up Vault →
            </Link>
            <button
              type="button"
              onClick={onCancel}
              className="lk-btn lk-btn--ghost w-full py-3"
            >
              Cancel
            </button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="flex flex-col gap-5">
            <div className="flex flex-col gap-2">
              <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
                Vault Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Combination"
                required
                autoFocus
                className="lk-input"
              />
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="lk-btn lk-btn--primary w-full py-3"
            >
              {isLoading ? "Turning dial..." : "Unlock Vault"}
            </button>

            <button
              type="button"
              onClick={onCancel}
              className="lk-btn lk-btn--ghost w-full py-3"
            >
              Cancel
            </button>
          </form>
        )}
      </div>
    </div>
  );
}