import { useState } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "~/providers/auth";
import { useToast } from "~/providers/toast";
import { buildKDFParams, deriveKeys, generateSalt } from "~/lib/crypto";

import { requireAuth } from "~/lib/auth-guard";
import type { Route } from "./+types/setup";
import { fetchKDFParams, setVerificationHash } from "~/lib/api/auth";

export const clientLoader = () => {
  requireAuth();
  return null;
};

export function meta({ }: Route.MetaArgs) {
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

  const calculateStrength = (pw: string): number => {
    let score = 0;
    if (pw.length >= 12) score += 1;
    if (pw.length >= 16) score += 1;
    if (/[A-Z]/.test(pw)) score += 1;
    if (/[0-9]/.test(pw)) score += 1;
    if (/[^A-Za-z0-9]/.test(pw)) score += 1;
    return score;
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const pw = e.target.value;
    setPassword(pw);
    setStrength(calculateStrength(pw));
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
      const _kdfParams = await fetchKDFParams()
      const salt = generateSalt()
      const kdfParams = buildKDFParams(_kdfParams, salt)
      const { encryptionKey, verificationHash } = await deriveKeys(password, kdfParams);

      const data = await setVerificationHash(verificationHash, kdfParams);

      login(data.user, accessToken);

      (window as any).__vaultKey = encryptionKey;

      addToast("Vault secured successfully", "success");
      navigate("/dashboard");
    } catch (err) {
      console.error({ ERRORLOUD: err })
      addToast("DOOM")
      // addToast(
      //   err instanceof Error ? err.message : "Failed to set vault password",
      //   "error",
      // );
    } finally {
      setIsLoading(false);
    }
  };

  const strengthLabels = [
    "Very Weak",
    "Weak",
    "Fair",
    "Good",
    "Strong",
    "Very Strong",
  ];
  const strengthColors = [
    "bg-red-500",
    "bg-red-400",
    "bg-amber-400",
    "bg-lime-400",
    "bg-green-400",
    "bg-emerald-500",
  ];

  return (
    <div className="flex min-h-[calc(100vh-8rem)] items-center justify-center px-4">
      <div className="w-full max-w-lg rounded-2xl border border-slate-800 bg-slate-900 p-8">
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
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-slate-100">
            Set Your Vault Password
          </h1>
          <p className="mt-2 text-sm text-slate-500">
            This is the only password you'll need to remember. We cannot reset
            it for you.
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
              onChange={handlePasswordChange}
              placeholder="Minimum 12 characters"
              required
              autoFocus
              className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-3 text-sm text-slate-200 placeholder:text-slate-600 focus:border-sky-400 focus:outline-none"
            />
            {password && (
              <div className="flex items-center gap-3">
                <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-slate-800">
                  <div
                    className={`h-full rounded-full transition-all duration-300 ${strengthColors[strength]}`}
                    style={{ width: `${(strength / 5) * 100}%` }}
                  />
                </div>
                <span className="text-xs font-medium text-slate-400">
                  {strengthLabels[strength]}
                </span>
              </div>
            )}
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-sm font-medium text-slate-400">
              Confirm Vault Password
            </label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Repeat your vault password"
              required
              className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-3 text-sm text-slate-200 placeholder:text-slate-600 focus:border-sky-400 focus:outline-none"
            />
          </div>

          <div className="flex gap-3 rounded-lg border border-amber-500/20 bg-amber-500/5 p-4 text-sm text-amber-300">
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
              If you lose this password, your vault is permanently inaccessible.
              There is no recovery mechanism.
            </span>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="rounded-lg bg-sky-400 py-3 text-sm font-semibold text-slate-950 hover:bg-sky-300 disabled:opacity-50 cursor-pointer"
          >
            {isLoading ? "Securing vault..." : "Create Vault"}
          </button>
        </form>
      </div>
    </div>
  );
}