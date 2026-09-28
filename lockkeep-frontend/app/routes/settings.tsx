import { useState } from "react";
import { useAuth } from "~/providers/auth";
import { useVault } from "~/providers/vault";
import { useToast } from "~/providers/toast";

import { calculatePasswordStrength } from "~/utils/calculatePasswordStrength";
import AccordionSection from "~/components/Accordion";
import Nameplate from "~/components/Nameplate";
import PasswordStrength from "~/components/PasswordStrength";
import { requireAuth } from "~/lib/auth-guard";
import type { Route } from "./+types/settings";
import { updateAccountPassword, updateEmail } from "~/lib/api/auth";
import { useNavigate } from "react-router";

export const clientLoader = () => {
  requireAuth();
  return null;
};

export function meta({}: Route.MetaArgs) {
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
  const { changeVaultPassword } = useVault();
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
      throw new Error("Not authenticated");
    }
    if (!user || isOAuth) return;

    setIsUpdatingEmail(true);
    try {
      await updateEmail(email);
      logout();
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
      await updateAccountPassword(currentAccountPassword, newAccountPassword);
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
      <div className="mb-8">
        <Nameplate>Operations</Nameplate>
        <h1 className="mt-2 font-display text-3xl font-semibold tracking-tight text-ivory">
          Settings
        </h1>
      </div>

      {/* Account Accordion */}
      <AccordionSection
        id="account"
        title="Account"
        isOpen={openSection === "account"}
        onToggle={() => toggleSection("account")}
      >
        {isOAuth ? (
          <div className="rounded-lg border border-brass-500/20 bg-brass-500/5 p-4">
            <p className="text-sm text-brass-300">
              Your account is managed by{" "}
              {user?.authMethod === "oauth_google" ? "Google" : "GitHub"} OAuth.
              Email and account password cannot be changed here.
            </p>
          </div>
        ) : (
          <div className="flex flex-col gap-6">
            <form onSubmit={handleEmailUpdate} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
                  Email Address
                </label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="lk-input"
                />
              </div>
              <div className="flex justify-end">
                <button
                  type="submit"
                  disabled={isUpdatingEmail || email === user?.email}
                  className="lk-btn lk-btn--primary"
                >
                  {isUpdatingEmail ? "Updating..." : "Update Email"}
                </button>
              </div>
            </form>

            <div className="h-px bg-brass-500/10" />

            <form
              onSubmit={handleAccountPasswordChange}
              className="flex flex-col gap-4"
            >
              <h3 className="text-sm font-semibold uppercase tracking-[0.14em] text-sand-400">
                Change Account Password
              </h3>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-sand-600">Current Password</label>
                <input
                  type="password"
                  value={currentAccountPassword}
                  onChange={(e) => setCurrentAccountPassword(e.target.value)}
                  required
                  className="lk-input"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-sand-600">New Password</label>
                <input
                  type="password"
                  value={newAccountPassword}
                  onChange={(e) => setNewAccountPassword(e.target.value)}
                  required
                  placeholder="At least 8 characters"
                  className="lk-input"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-sand-600">
                  Confirm New Password
                </label>
                <input
                  type="password"
                  value={confirmAccountPassword}
                  onChange={(e) => setConfirmAccountPassword(e.target.value)}
                  required
                  className="lk-input"
                />
              </div>
              <div className="flex justify-end">
                <button
                  type="submit"
                  disabled={isChangingAccountPassword}
                  className="lk-btn lk-btn--primary"
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
            <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
              Current Vault Password
            </label>
            <input
              type="password"
              value={currentVaultPassword}
              onChange={(e) => setCurrentVaultPassword(e.target.value)}
              required
              className="lk-input"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
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
              className="lk-input"
            />
            <PasswordStrength
              password={newVaultPassword}
              score={masterPasswordStrength}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
              Confirm New Vault Password
            </label>
            <input
              type="password"
              value={confirmVaultPassword}
              onChange={(e) => setConfirmVaultPassword(e.target.value)}
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
              Changing your vault password will re-encrypt vault content with a
              new key. This operation may take a moment depending on how many
              items you have.
            </span>
          </div>

          <div className="flex justify-end">
            <button
              type="submit"
              disabled={isChangingVaultPassword}
              className="lk-btn lk-btn--primary"
            >
              {isChangingVaultPassword
                ? "Re-encrypting..."
                : "Change Vault Password"}
            </button>
          </div>
        </form>
      </AccordionSection>
    </div>
  );
}