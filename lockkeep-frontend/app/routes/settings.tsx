import { useState } from "react";
import { useAuth } from "~/providers/auth";
import { useVault } from "~/providers/vault";
import { useToast } from "~/providers/toast";

import {
  calculatePasswordStrength,
  strengthColors,
  strengthLabels,
} from "~/utils/calculatePasswordStrength";
import AccordionSection from "~/components/Accordion";
import { requireAuth } from "~/lib/auth-guard";
import type { Route } from "./+types/settings";
import { updateAccountPassword, updateEmail } from "~/lib/api/auth";
import { useNavigate } from "react-router";

export const clientLoader = () => {
  requireAuth();
  return null;
};

export function meta({ }: Route.MetaArgs) {
  return [
    { title: "Settings - LockKeep" },
    {
      name: "description",
      content:
        "Manage your LockKeep account settings, change your vault password, and update security preferences.",
    },
    { name: "robots", content: "noindex, nofollow" },
  ];
}

export default function Settings() {
  const { user, getAccessToken, logout } = useAuth();
  const { isLocked, changeVaultPassword } = useVault();
  const { addToast } = useToast();

  const [openSection, setOpenSection] = useState<string | null>("account");

  const [email, setEmail] = useState(user?.email || "");
  const [isUpdatingEmail, setIsUpdatingEmail] = useState(false);

  const [currentAccountPassword, setCurrentAccountPassword] = useState("");
  const [newAccountPassword, setNewAccountPassword] = useState("");
  const [confirmAccountPassword, setConfirmAccountPassword] = useState("");
  const [isChangingAccountPassword, setIsChangingAccountPassword] =
    useState(false);

  const [currentVaultPassword, setCurrentVaultPassword] = useState("");
  const [newVaultPassword, setNewVaultPassword] = useState("");
  const [confirmVaultPassword, setConfirmVaultPassword] = useState("");
  const [isChangingVaultPassword, setIsChangingVaultPassword] =
    useState(false);
  const [masterPasswordStrength, setVaultPasswordStrength] = useState(0);

  const navigate = useNavigate();

  const isOAuth = user?.authMethod !== "local";

  const accessToken = getAccessToken();

  const handleEmailUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!accessToken) {
      throw new Error("Not authenticated")
    };
    if (!user || isOAuth) return;

    setIsUpdatingEmail(true);
    try {
      await updateEmail(email);
      logout()
      navigate("/login");
      addToast("Email updated successfully", "success");
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to update email",
        "error",
      );
    } finally {
      setIsUpdatingEmail(false);
    }
  };

  const handleAccountPasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();

    if (newAccountPassword !== confirmAccountPassword) {
      addToast("Passwords do not match", "error");
      return;
    }

    if (newAccountPassword.length < 8) {
      addToast("Password must be at least 8 characters", "error");
      return;
    }

    setIsChangingAccountPassword(true);
    try {
      await updateAccountPassword(
        currentAccountPassword,
        newAccountPassword,
      );
      addToast("Account password changed successfully", "success");
      setCurrentAccountPassword("");
      setNewAccountPassword("");
      setConfirmAccountPassword("");
    } catch (err) {
      addToast(
        err instanceof Error
          ? err.message
          : "Failed to change account password",
        "error",
      );
    } finally {
      setIsChangingAccountPassword(false);
    }
  };

  const handleVaultPasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();

    if (newVaultPassword.length < 12) {
      addToast("New vault password must be at least 12 characters", "error");
      return;
    }

    if (newVaultPassword !== confirmVaultPassword) {
      addToast("New vault passwords do not match", "error");
      return;
    }

    if (masterPasswordStrength < 3) {
      addToast("Please use a stronger vault password", "error");
      return;
    }

    // if (isLocked) {
    //   addToast("Vault must be unlocked to change vault password", "error");
    //   return;
    // }

    setIsChangingVaultPassword(true);
    try {
      await changeVaultPassword(currentVaultPassword, newVaultPassword);
      addToast(
        "Vault password changed successfully. All credentials re-encrypted.",
        "success",
      );
      setCurrentVaultPassword("");
      setNewVaultPassword("");
      setConfirmVaultPassword("");
      setVaultPasswordStrength(0);
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to change vault password",
        "error",
      );
    } finally {
      setIsChangingVaultPassword(false);
    }
  };

  const toggleSection = (section: string) => {
    setOpenSection(openSection === section ? null : section);
  };

  return (
    <div className="mx-auto max-w-2xl">
      <h1 className="mb-8 text-3xl font-bold text-slate-100">Settings</h1>

      {/* Account Accordion */}
      <AccordionSection
        id="account"
        title="Account"
        isOpen={openSection === "account"}
        onToggle={() => toggleSection("account")}
      >
        {isOAuth ? (
          <div className="rounded-lg border border-sky-500/20 bg-sky-500/5 p-4">
            <p className="text-sm text-sky-300">
              Your account is managed by{" "}
              {user?.authMethod === "oauth_google" ? "Google" : "GitHub"} OAuth.
              Email and account password cannot be changed here.
            </p>
          </div>
        ) : (
          <div className="flex flex-col gap-6">
            <form onSubmit={handleEmailUpdate} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium text-slate-400">
                  Email Address
                </label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-sm text-slate-200 focus:border-sky-400 focus:outline-none"
                />
              </div>
              <div className="flex justify-end">
                <button
                  type="submit"
                  disabled={isUpdatingEmail || email === user?.email}
                  className="rounded-lg bg-sky-400 px-4 py-2 text-sm font-semibold text-slate-950 hover:bg-sky-300 disabled:opacity-50"
                >
                  {isUpdatingEmail ? "Updating..." : "Update Email"}
                </button>
              </div>
            </form>

            <div className="h-px bg-slate-800" />

            <form
              onSubmit={handleAccountPasswordChange}
              className="flex flex-col gap-4"
            >
              <h3 className="text-sm font-medium text-slate-400">
                Change Account Password
              </h3>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-slate-500">
                  Current Password
                </label>
                <input
                  type="password"
                  value={currentAccountPassword}
                  onChange={(e) => setCurrentAccountPassword(e.target.value)}
                  required
                  className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-sm text-slate-200 focus:border-sky-400 focus:outline-none"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-slate-500">New Password</label>
                <input
                  type="password"
                  value={newAccountPassword}
                  onChange={(e) => setNewAccountPassword(e.target.value)}
                  required
                  placeholder="At least 8 characters"
                  className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-sm text-slate-200 placeholder:text-slate-600 focus:border-sky-400 focus:outline-none"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-slate-500">
                  Confirm New Password
                </label>
                <input
                  type="password"
                  value={confirmAccountPassword}
                  onChange={(e) => setConfirmAccountPassword(e.target.value)}
                  required
                  className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-sm text-slate-200 focus:border-sky-400 focus:outline-none"
                />
              </div>
              <div className="flex justify-end">
                <button
                  type="submit"
                  disabled={isChangingAccountPassword}
                  className="rounded-lg bg-sky-400 px-4 py-2 text-sm font-semibold text-slate-950 hover:bg-sky-300 disabled:opacity-50"
                >
                  {isChangingAccountPassword
                    ? "Updating..."
                    : "Change Password"}
                </button>
              </div>
            </form>
          </div>
        )}
      </AccordionSection>

      {/* Vault Security Accordion */}
      <AccordionSection
        id="vault"
        title="Vault Security"
        isOpen={openSection === "vault"}
        onToggle={() => toggleSection("vault")}
      >
        <form
          onSubmit={handleVaultPasswordChange}
          className="flex flex-col gap-4"
        >
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-slate-400">
              Current Vault Password
            </label>
            <input
              type="password"
              value={currentVaultPassword}
              onChange={(e) => setCurrentVaultPassword(e.target.value)}
              required
              className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-sm text-slate-200 focus:border-sky-400 focus:outline-none"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-slate-400">
              New Vault Password
            </label>
            <input
              type="password"
              value={newVaultPassword}
              onChange={(e) => {
                setNewVaultPassword(e.target.value);
                setVaultPasswordStrength(
                  calculatePasswordStrength(e.target.value),
                );
              }}
              required
              placeholder="Minimum 12 characters"
              className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-sm text-slate-200 placeholder:text-slate-600 focus:border-sky-400 focus:outline-none"
            />
            {newVaultPassword && (
              <div className="flex items-center gap-3">
                <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-slate-800">
                  <div
                    className={`h-full rounded-full transition-all duration-300 ${strengthColors[masterPasswordStrength]}`}
                    style={{ width: `${(masterPasswordStrength / 5) * 100}%` }}
                  />
                </div>
                <span className="text-xs font-medium text-slate-400">
                  {strengthLabels[masterPasswordStrength]}
                </span>
              </div>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-slate-400">
              Confirm New Vault Password
            </label>
            <input
              type="password"
              value={confirmVaultPassword}
              onChange={(e) => setConfirmVaultPassword(e.target.value)}
              required
              className="rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-sm text-slate-200 focus:border-sky-400 focus:outline-none"
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
              Changing your vault password will re-encrypt vault content with
              a new key. This operation may take a moment depending on how many
              items you have.
            </span>
          </div>

          <div className="flex justify-end">
            <button
              type="submit"
              disabled={isChangingVaultPassword}
              className="rounded-lg bg-sky-400 px-4 py-2 text-sm font-semibold text-slate-950 hover:bg-sky-300 disabled:opacity-50"
            >
              {isChangingVaultPassword
                ? "Updating..."
                : "Change Vault Password"}
            </button>
          </div>
        </form>
      </AccordionSection>
    </div>
  );
}
