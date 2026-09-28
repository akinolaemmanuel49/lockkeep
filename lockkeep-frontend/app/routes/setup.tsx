import { useState } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "~/providers/auth";
import type { User } from "~/types";
import { useToast } from "~/providers/toast";
import { buildKDFParams, deriveKeys, generateSalt } from "~/lib/crypto";
import { calculatePasswordStrength } from "~/utils/calculatePasswordStrength";
import VaultDial from "~/components/VaultDial";
import Nameplate from "~/components/Nameplate";
import PasswordStrength from "~/components/PasswordStrength";

import { requireAuth } from "~/lib/auth-guard";
import type { Route } from "./+types/setup";
import { fetchKDFParams, setVerificationHash } from "~/lib/api/auth";

export const clientLoader = () => {
  requireAuth();
  return null;
};

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Set Up Vault - LockKeep" },
    {
      name: "description",
      content:
        "Set your vault password to secure your LockKeep vault with zero-trust encryption.",
    },
    { name: "robots", content: "noindex, nofollow" },
  ];
}

export default function Setup() {
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [strength, setStrength] = useState(0);
  const { user, login, getAccessToken } = useAuth();
  const { addToast } = useToast();
  const navigate = useNavigate();

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const pw = e.target.value;
    setPassword(pw);
    setStrength(calculatePasswordStrength(pw));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (password.length < 12) {
      addToast("Vault password must be at least 12 characters", "error");
      return;
    }

    if (password !== confirmPassword) {
      addToast("Passwords do not match", "error");
      return;
    }

    if (strength < 3) {
      addToast("Please use a stronger password", "error");
      return;
    }

    if (!user) {
      addToast("Not authenticated", "error");
      return;
    }

    const accessToken = getAccessToken();
    if (!accessToken) {
      addToast("Session expired. Please sign in again.", "error");
      navigate("/login");
      return;
    }

    setIsLoading(true);
    try {
      const _kdfParams = await fetchKDFParams();
      const salt = generateSalt();
      const kdfParams = buildKDFParams(_kdfParams, salt);
      const { encryptionKey, verificationHash } = await deriveKeys(
        password,
        kdfParams,
      );

      const data = await setVerificationHash(verificationHash, kdfParams);

      login(
        {
          ...(user as User),
          ...data.user,
          hasMasterPassword: true,
        },
        accessToken,
      );

      (window as any).__vaultKey = encryptionKey;

      addToast("Vault secured successfully", "success");
      navigate("/dashboard");
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to set vault password",
        "error",
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex min-h-[calc(100vh-9rem)] items-center justify-center px-4">
      <div className="relative w-full max-w-lg">
        <VaultDial
          size={460}
          className="pointer-events-none absolute -left-32 -bottom-24 opacity-[0.1]"
        />
        <div className="plate relative p-8">
          <div className="mb-8 text-center">
            <VaultDial size={64} className="mx-auto text-brass-400" />
            <Nameplate className="mt-5 justify-center">
              First Seal
            </Nameplate>
            <h1 className="mt-3 font-display text-2xl font-semibold tracking-tight text-ivory">
              Set Your Vault Password
            </h1>
            <p className="mt-2 text-sm text-sand-500">
              The only password you'll need to remember. We cannot reset it for
              you — ever.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="flex flex-col gap-5">
            <div className="flex flex-col gap-2">
              <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
                Vault Password
              </label>
              <input
                type="password"
                value={password}
                onChange={handlePasswordChange}
                placeholder="Minimum 12 characters"
                required
                autoFocus
                className="lk-input"
              />
              <PasswordStrength password={password} score={strength} />
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
                Confirm Vault Password
              </label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Repeat your vault password"
                required
                className="lk-input"
              />
            </div>

            <div className="flex gap-3 rounded-lg border border-vermillion-500/25 bg-vermillion-500/5 p-4 text-sm text-vermillion-300">
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="mt-0.5 shrink-0"
              >
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                <line x1="12" y1="9" x2="12" y2="13" />
                <line x1="12" y1="17" x2="12.01" y2="17" />
              </svg>
              <span>
                If you lose this password, your vault is permanently
                inaccessible. There is no recovery mechanism.
              </span>
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="lk-btn lk-btn--primary w-full py-3"
            >
              {isLoading ? "Sealing vault..." : "Create Vault"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}