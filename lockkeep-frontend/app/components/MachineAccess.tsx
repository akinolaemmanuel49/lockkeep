import { useEffect, useState } from "react";
import type {
    ApiKeyRecord,
    DecryptedMachineSecret,
    ServiceAccount,
    WorkspaceApplication,
    WorkspaceEnvironment,
} from "~/types";
import {
    createServiceAccount,
    deleteServiceAccount,
    fetchApiKeys,
    fetchServiceAccounts,
    mintApiKey,
    revokeApiKey,
} from "~/lib/api/serviceAccounts";
import { fetchAndDecryptEnvelope } from "~/lib/api/machine";
import { useToast } from "~/providers/toast";
import Nameplate from "./Nameplate";

interface MachineAccessProps {
    orgSlug: string;
    app: WorkspaceApplication;
    env: WorkspaceEnvironment | null;
}

function KeyRow({
    label,
    value,
    onCopy,
    actions,
}: {
    label: string;
    value: string;
    onCopy?: () => void;
    actions?: React.ReactNode;
}) {
    return (
        <div className="rounded-md border border-brass-500/15 bg-coal-950 px-3 py-2.5">
            <div className="flex items-center justify-between gap-2">
                <span className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                    {label}
                </span>
                <div className="flex items-center gap-3">
                    {actions}
                    {onCopy && (
                        <button
                            type="button"
                            onClick={onCopy}
                            className="text-xs text-sand-500 transition-colors hover:text-brass-300"
                        >
                            Copy
                        </button>
                    )}
                </div>
            </div>
            <p className="mt-1.5 break-all font-mono text-xs text-ivory">{value}</p>
        </div>
    );
}

export default function MachineAccess({
    orgSlug,
    app,
    env,
}: MachineAccessProps) {
    const { addToast } = useToast();

    const [accounts, setAccounts] = useState<ServiceAccount[]>([]);
    const [accountsError, setAccountsError] = useState<string | null>(null);

    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [accountName, setAccountName] = useState("");
    const [accountDescription, setAccountDescription] = useState("");
    const [isCreating, setIsCreating] = useState(false);

    const [expandedId, setExpandedId] = useState<string | null>(null);
    const [keysByAccount, setKeysByAccount] = useState<
        Record<string, ApiKeyRecord[]>
    >({});
    const [keysLoadingId, setKeysLoadingId] = useState<string | null>(null);
    const [keysError, setKeysError] = useState<string | null>(null);

    const [mintingId, setMintingId] = useState<string | null>(null);
    const [expiryDays, setExpiryDays] = useState<number>(0);

    const [rawKey, setRawKey] = useState<string | null>(null);
    const [rawKeyNote, setRawKeyNote] = useState<string | null>(null);

    const [preview, setPreview] = useState<DecryptedMachineSecret[] | null>(
        null,
    );
    const [previewEnvId, setPreviewEnvId] = useState<string | null>(null);
    const [revealed, setRevealed] = useState<Record<string, boolean>>({});
    const [isDecrypting, setIsDecrypting] = useState(false);
    const [decryptError, setDecryptError] = useState<string | null>(null);

    useEffect(() => {
        let cancelled = false;
        (async () => {
            setAccountsError(null);
            try {
                const items = await fetchServiceAccounts(orgSlug);
                if (cancelled) return;
                setAccounts(items);
            } catch (err) {
                if (!cancelled) {
                    setAccountsError(
                        err instanceof Error
                            ? err.message
                            : "Failed to load service accounts",
                    );
                }
            }
        })();
        return () => {
            cancelled = true;
        };
    }, [orgSlug]);

    // A minted key is bound to the selected environment; reset when it changes.
    useEffect(() => {
        setRawKey(null);
        setRawKeyNote(null);
        setPreview(null);
        setPreviewEnvId(null);
        setRevealed({});
        setDecryptError(null);
    }, [env?.slug]);

    const loadKeys = async (accountId: string) => {
        setKeysLoadingId(accountId);
        setKeysError(null);
        try {
            const keys = await fetchApiKeys(orgSlug, accountId);
            setKeysByAccount((m) => ({ ...m, [accountId]: keys }));
        } catch (err) {
            setKeysError(
                err instanceof Error ? err.message : "Failed to load keys",
            );
        } finally {
            setKeysLoadingId(null);
        }
    };

    const toggleExpand = (accountId: string) => {
        if (expandedId === accountId) {
            setExpandedId(null);
            return;
        }
        setExpandedId(accountId);
        if (!keysByAccount[accountId]) void loadKeys(accountId);
    };

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!accountName.trim()) return;
        setIsCreating(true);
        try {
            const created = await createServiceAccount(orgSlug, {
                name: accountName.trim(),
                description: accountDescription.trim() || undefined,
            });
            addToast("Service account created", "success");
            setAccounts((items) => [...items, created]);
            setAccountName("");
            setAccountDescription("");
            setIsCreateOpen(false);
        } catch (err) {
            addToast(
                err instanceof Error ? err.message : "Failed to create account",
                "error",
            );
        } finally {
            setIsCreating(false);
        }
    };

    const handleDeleteAccount = async (account: ServiceAccount) => {
        const confirmed = window.confirm(
            `Delete "${account.name}"? All of its API keys are revoked immediately.`,
        );
        if (!confirmed) return;
        try {
            await deleteServiceAccount(orgSlug, account.id);
            addToast("Service account deleted", "success");
            setAccounts((items) => items.filter((a) => a.id !== account.id));
            if (expandedId === account.id) setExpandedId(null);
        } catch (err) {
            addToast(
                err instanceof Error
                    ? err.message
                    : "Failed to delete account",
                "error",
            );
        }
    };

    const handleMint = async (account: ServiceAccount) => {
        if (!env) return;
        setMintingId(account.id);
        setRawKeyNote(null);
        try {
            const { apiKey } = await mintApiKey(orgSlug, account.id, {
                applicationSlug: app.slug,
                environmentSlug: env.slug,
                scopes: ["secrets:read"],
                expiresAt:
                    expiryDays > 0
                        ? new Date(
                              Date.now() + expiryDays * 24 * 60 * 60 * 1000,
                          ).toISOString()
                        : undefined,
            });
            setRawKey(apiKey);
            setRawKeyNote(`granted ${app.slug} · ${env.slug}`);
            addToast(
                "API key minted — copy it now; it will not be shown again.",
                "success",
            );
            void loadKeys(account.id);
        } catch (err) {
            addToast(
                err instanceof Error ? err.message : "Failed to mint API key",
                "error",
            );
        } finally {
            setMintingId(null);
        }
    };

    const handleRevoke = async (account: ServiceAccount, key: ApiKeyRecord) => {
        const confirmed = window.confirm(
            `Revoke ${key.keyPrefix}…? Machines using it lose access immediately.`,
        );
        if (!confirmed) return;
        try {
            await revokeApiKey(orgSlug, account.id, key.id);
            addToast("API key revoked", "success");
            await loadKeys(account.id);
        } catch (err) {
            addToast(
                err instanceof Error ? err.message : "Failed to revoke key",
                "error",
            );
        }
    };

    const handleDecryptPreview = async () => {
        if (!rawKey) return;
        setIsDecrypting(true);
        setDecryptError(null);
        try {
            const result = await fetchAndDecryptEnvelope(rawKey);
            setPreview(result.secrets);
            setPreviewEnvId(result.environmentId);
        } catch (err) {
            setDecryptError(
                err instanceof Error
                    ? err.message
                    : "Failed to decrypt environment",
            );
        } finally {
            setIsDecrypting(false);
        }
    };

    const copyText = async (text: string) => {
        try {
            await navigator.clipboard.writeText(text);
            addToast("Copied to clipboard", "success");
        } catch {
            addToast("Copy failed — select and copy manually", "error");
        }
    };

    return (
        <section className="plate p-6">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <Nameplate>Machine Access</Nameplate>
                <span className="font-mono text-xs text-sand-500">
                    service accounts → environment keys
                </span>
            </div>

            <p className="mt-3 max-w-3xl text-sm leading-relaxed text-sand-500">
                {env
                    ? `Keys minted here are bound to ${env.name} (${app.slug} · ${env.slug}) and can decrypt its secrets — the server only returns ciphertext, so the key never leaks values in transit.`
                    : "Select an environment above to mint environment-scoped keys for it."}
            </p>

            {rawKey && (
                <div className="mt-4 rounded-lg border border-brass-500/25 bg-brass-500/5 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                        <span className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-brass-300">
                            New key — shown once
                        </span>
                        <button
                            type="button"
                            onClick={() => {
                                setRawKey(null);
                                setPreview(null);
                            }}
                            className="text-xs text-sand-600 transition-colors hover:text-sand-400"
                        >
                            Dismiss
                        </button>
                    </div>
                    <KeyRow
                        label={rawKeyNote ?? "raw key"}
                        value={rawKey}
                        onCopy={() => void copyText(rawKey)}
                        actions={
                            <button
                                type="button"
                                onClick={() => void handleDecryptPreview()}
                                disabled={isDecrypting}
                                className="lk-btn lk-btn--primary !px-2.5 !py-1 text-xs"
                            >
                                {isDecrypting
                                    ? "Decrypting…"
                                    : "Fetch & decrypt this environment"}
                            </button>
                        }
                    />

                    {decryptError && (
                        <p className="mt-2.5 rounded-md border border-vermillion-500/25 bg-vermillion-500/10 px-3 py-2 text-xs text-vermillion-300">
                            {decryptError}
                        </p>
                    )}

                    {preview && (
                        <div className="mt-4">
                            <div className="flex items-center justify-between gap-2">
                                <span className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                                    In-browser decrypt preview
                                </span>
                                <span className="font-mono text-[0.625rem] text-sand-600">
                                    env {previewEnvId}
                                </span>
                            </div>
                            <div className="mt-2 flex flex-col gap-2">
                                {preview.length === 0 ? (
                                    <p className="rounded-md border border-dashed border-brass-500/20 px-3 py-4 text-center text-xs text-sand-600">
                                        No secrets in this environment yet.
                                    </p>
                                ) : (
                                    preview.map((secret) => (
                                        <div
                                            key={secret.key}
                                            className="rounded-md border border-brass-500/15 bg-coal-950 p-3"
                                        >
                                            <div className="flex items-center justify-between gap-2">
                                                <div className="flex min-w-0 items-center gap-2">
                                                    <p className="truncate font-mono text-sm font-medium text-ivory">
                                                        {secret.key}
                                                    </p>
                                                    <span className="rounded-full border border-patina-500/25 bg-patina-500/10 px-2 py-0.5 font-mono text-[0.625rem] uppercase tracking-[0.14em] text-patina-400">
                                                        {secret.kind}
                                                    </span>
                                                    <span className="shrink-0 font-mono text-[0.625rem] text-sand-600">
                                                        v{secret.version}
                                                    </span>
                                                </div>
                                                <button
                                                    type="button"
                                                    onClick={() =>
                                                        setRevealed((r) => ({
                                                            ...r,
                                                            [secret.key]:
                                                                !r[secret.key],
                                                        }))
                                                    }
                                                    className="text-xs text-sand-500 transition-colors hover:text-brass-300"
                                                >
                                                    {revealed[secret.key]
                                                        ? "Hide"
                                                        : "Show"}
                                                </button>
                                            </div>
                                            <pre className="mt-2 overflow-x-auto whitespace-pre-wrap break-all rounded-md border border-brass-500/10 bg-coal-900 px-3 py-2 font-mono text-xs text-patina-300">
                                                {revealed[secret.key]
                                                    ? secret.value || "(empty)"
                                                    : "••••••••••••••••"}
                                            </pre>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>
                    )}
                </div>
            )}

            {/* Service accounts */}
            <div className="mt-6 grid grid-cols-1 gap-5 lg:grid-cols-2">
                <div className="flex flex-col gap-3">
                    <div className="flex items-center justify-between gap-2">
                        <h3 className="text-xs font-semibold uppercase tracking-[0.2em] text-sand-500">
                            Service Accounts
                        </h3>
                        <button
                            type="button"
                            onClick={() => setIsCreateOpen((v) => !v)}
                            className="lk-btn lk-btn--ghost !px-2.5 !py-1 text-xs"
                        >
                            {isCreateOpen ? "Close" : "+ New Account"}
                        </button>
                    </div>

                    {isCreateOpen && (
                        <form
                            onSubmit={handleCreate}
                            className="flex flex-col gap-2.5 rounded-lg border border-brass-500/15 bg-coal-950 p-4"
                        >
                            <div className="flex flex-col gap-1.5">
                                <label className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                                    Name *
                                </label>
                                <input
                                    type="text"
                                    value={accountName}
                                    onChange={(e) =>
                                        setAccountName(e.target.value)
                                    }
                                    placeholder="e.g., CI Runner"
                                    className="lk-input"
                                />
                            </div>
                            <div className="flex flex-col gap-1.5">
                                <label className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                                    Description
                                </label>
                                <input
                                    type="text"
                                    value={accountDescription}
                                    onChange={(e) =>
                                        setAccountDescription(e.target.value)
                                    }
                                    placeholder="e.g., GitHub Actions deploy job"
                                    className="lk-input"
                                />
                            </div>
                            <button
                                type="submit"
                                disabled={isCreating}
                                className="lk-btn lk-btn--primary"
                            >
                                {isCreating ? "Creating…" : "Create Account"}
                            </button>
                        </form>
                    )}

                    {accountsError && (
                        <p className="rounded-lg border border-vermillion-500/25 bg-vermillion-500/10 px-3 py-2 text-xs text-vermillion-300">
                            {accountsError}
                        </p>
                    )}

                    {accounts.length === 0 && !accountsError && (
                        <p className="rounded-lg border border-dashed border-brass-500/20 px-3 py-5 text-center text-xs text-sand-600">
                            No service accounts yet. Create one to mint API keys.
                        </p>
                    )}

                    <div className="flex flex-col gap-2">
                        {accounts.map((account) => (
                            <div
                                key={account.id}
                                className="rounded-lg border border-brass-500/15 bg-coal-950 px-3 py-2.5"
                            >
                                <div className="flex items-center justify-between gap-2">
                                    <button
                                        type="button"
                                        onClick={() =>
                                            toggleExpand(account.id)
                                        }
                                        className="min-w-0 flex-1 text-left"
                                    >
                                        <p className="truncate text-sm font-medium text-ivory">
                                            {account.name}
                                            {account.disabledAt && (
                                                <span className="ml-2 text-vermillion-300">
                                                    · disabled
                                                </span>
                                            )}
                                        </p>
                                        <p className="truncate font-mono text-xs text-sand-500">
                                            {account.description || account.id}
                                        </p>
                                    </button>
                                    <div className="flex shrink-0 items-center gap-3 text-xs">
                                        <button
                                            type="button"
                                            disabled={!env || mintingId !== null}
                                            onClick={() =>
                                                void handleMint(account)
                                            }
                                            className="text-brass-300 transition-colors hover:text-brass-200 disabled:cursor-not-allowed disabled:opacity-40"
                                        >
                                            {mintingId === account.id
                                                ? "Minting…"
                                                : "Mint Key"}
                                        </button>
                                        <button
                                            type="button"
                                            onClick={() =>
                                                void handleDeleteAccount(account)
                                            }
                                            className="text-sand-600 transition-colors hover:text-vermillion-300"
                                        >
                                            Delete
                                        </button>
                                    </div>
                                </div>

                                {expandedId === account.id && (
                                    <div className="mt-3 flex flex-col gap-2 border-t border-brass-500/10 pt-3">
                                        {keysLoadingId === account.id ? (
                                            <p className="text-xs text-sand-600">
                                                Loading keys…
                                            </p>
                                        ) : keysError ? (
                                            <p className="text-xs text-vermillion-300">
                                                {keysError}
                                            </p>
                                        ) : (
                                            <>
                                                <ExpirySelect
                                                    value={expiryDays}
                                                    onChange={setExpiryDays}
                                                />
                                                {(keysByAccount[
                                                    account.id
                                                ] ?? []).length === 0 ? (
                                                    <p className="rounded-md border border-dashed border-brass-500/20 px-3 py-3 text-center text-xs text-sand-600">
                                                        No keys yet. Mint one for
                                                        the selected environment.
                                                    </p>
                                                ) : (
                                                    (keysByAccount[
                                                        account.id
                                                    ] ?? []).map((key) => (
                                                        <KeyRow
                                                            key={key.id}
                                                            label={`${key.keyPrefix} · ${key.scopes.join(", ")}`}
                                                            value={`created ${new Date(
                                                                key.createdAt,
                                                            ).toLocaleString()}${key.expiresAt ? ` · expires ${new Date(key.expiresAt).toLocaleDateString()}` : ""}${key.lastUsedAt ? ` · last used ${new Date(key.lastUsedAt).toLocaleString()}` : ""}`}
                                                            actions={
                                                                !key.revokedAt && (
                                                                    <button
                                                                        type="button"
                                                                        onClick={() =>
                                                                            void handleRevoke(
                                                                                account,
                                                                                key,
                                                                            )
                                                                        }
                                                                        className="text-xs text-vermillion-300 transition-colors hover:text-vermillion-200"
                                                                    >
                                                                        Revoke
                                                                    </button>
                                                                )
                                                            }
                                                        />
                                                    ))
                                                )}
                                            </>
                                        )}
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                </div>

                <div className="flex flex-col gap-3">
                    <h3 className="text-xs font-semibold uppercase tracking-[0.2em] text-sand-500">
                        How it works
                    </h3>
                    <ol className="flex list-decimal flex-col gap-2.5 rounded-lg border border-brass-500/15 bg-coal-950 p-4 pl-8 text-sm leading-relaxed text-sand-400">
                        <li>
                            The key is minted against a single environment
                            (<span className="font-mono text-brass-300">
                                {env ? `${app.slug} · ${env.slug}` : "pick one above"}
                            </span>
                            ) with read-only scope.
                        </li>
                        <li>
                            A machine calls{" "}
                            <span className="font-mono text-xs text-sand-300">
                                GET /api/v2/machine/secrets
                            </span>{" "}
                            with the key; the server returns ciphertext only.
                        </li>
                        <li>
                            The client derives the wrapping key (HKDF-SHA256),
                            unwraps the environment data key (AES-256-GCM), and
                            decrypts each value locally. The{" "}
                            <span className="font-mono text-sand-300">
                                @lockkeep/sdk
                            </span>{" "}
                            automates this in CI.
                        </li>
                    </ol>
                    <p className="rounded-lg border border-brass-500/15 bg-coal-950 p-4 text-xs leading-relaxed text-sand-500">
                        Keyboard verification happens entirely in your browser —
                        the raw key never leaves this page. Every export is
                        written to the audit log with{" "}
                        <span className="font-mono text-patina-300">
                            action: export, decrypted: true
                        </span>
                        .
                    </p>
                </div>
            </div>
        </section>
    );
}

function ExpirySelect({
    value,
    onChange,
}: {
    value: number;
    onChange: (days: number) => void;
}) {
    return (
        <div className="flex items-center justify-between gap-2 rounded-md border border-brass-500/15 bg-coal-900 px-3 py-2">
            <label className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                Mint with expiration
            </label>
            <select
                value={value}
                onChange={(e) => onChange(Number(e.target.value))}
                className="lk-input !w-auto !py-1 font-mono text-xs"
            >
                <option value={0}>never</option>
                <option value={30}>30 days</option>
                <option value={90}>90 days</option>
                <option value={365}>1 year</option>
            </select>
        </div>
    );
}