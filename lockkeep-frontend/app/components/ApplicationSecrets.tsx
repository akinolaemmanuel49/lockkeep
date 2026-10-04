import { useEffect, useRef, useState } from "react";
import type {
    CreateEnvironmentInput,
    SharedSecretKind,
    WorkspaceApplication,
    WorkspaceEnvironment,
    WorkspaceEnvSecret,
} from "~/types";
import {
    createEnvironment,
    createSecret,
    deleteEnvironment,
    deleteSecret,
    fetchEnvironments,
    fetchSecrets,
    updateSecret,
} from "~/lib/api/sharedSecrets";
import { parseDotEnv } from "~/lib/parseDotEnv";
import { useToast } from "~/providers/toast";
import Nameplate from "./Nameplate";
import MachineAccess from "./MachineAccess";

const slugify = (value: string) =>
    value
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "");

const KIND_META: Record<SharedSecretKind, { tone: string }> = {
    env: { tone: "text-patina-400 border-patina-500/25 bg-patina-500/10" },
    note: { tone: "text-sand-400 border-sand-500/25 bg-sand-500/10" },
    file: { tone: "text-brass-400 border-brass-500/25 bg-brass-500/10" },
    json: { tone: "text-vermillion-400 border-vermillion-500/25 bg-vermillion-500/10" },
};

function KindBadge({ kind }: { kind: SharedSecretKind }) {
    return (
        <span
            className={`inline-flex rounded-full border px-2 py-0.5 font-mono text-[0.625rem] uppercase tracking-[0.14em] ${KIND_META[kind].tone}`}
        >
            {kind}
        </span>
    );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
    return (
        <div className="flex flex-col gap-1.5">
            <label className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                {label}
            </label>
            {children}
        </div>
    );
}

interface DraftSecretRow {
    key: string;
    kind: SharedSecretKind;
    value: string;
}

interface ApplicationSecretsProps {
    orgSlug: string;
    app: WorkspaceApplication;
}

export default function ApplicationSecrets({
    orgSlug,
    app,
}: ApplicationSecretsProps) {
    const { addToast } = useToast();

    // Environments level
    const [environments, setEnvironments] = useState<WorkspaceEnvironment[]>([]);
    const [selectedEnvSlug, setSelectedEnvSlug] = useState<string | null>(null);
    const [isEnvFormOpen, setIsEnvFormOpen] = useState(false);
    const [envForm, setEnvForm] = useState<CreateEnvironmentInput>({
        name: "",
        slug: "",
        isProtected: false,
    });
    const [envSlugTouched, setEnvSlugTouched] = useState(false);
    const [isSavingEnv, setIsSavingEnv] = useState(false);
    const [envError, setEnvError] = useState<string | null>(null);

    // Secrets level
    const [secrets, setSecrets] = useState<WorkspaceEnvSecret[]>([]);
    const [revealed, setRevealed] = useState<Record<string, boolean>>({});
    const [editingKey, setEditingKey] = useState<string | null>(null);
    const [isSecretFormOpen, setIsSecretFormOpen] = useState(false);
    const [secretForm, setSecretForm] = useState({
        key: "",
        kind: "env" as SharedSecretKind,
        value: "",
    });
    const [isSavingSecret, setIsSavingSecret] = useState(false);
    const [secretError, setSecretError] = useState<string | null>(null);

    // Batch import / replication (UI-only: copies keys/kinds, never values)
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [isCopyPickerOpen, setIsCopyPickerOpen] = useState(false);
    const [draftRows, setDraftRows] = useState<DraftSecretRow[]>([]);
    const [draftSourceLabel, setDraftSourceLabel] = useState<string | null>(
        null,
    );
    const [draftSkipped, setDraftSkipped] = useState(0);
    const [isBatchSaving, setIsBatchSaving] = useState(false);
    const [batchError, setBatchError] = useState<string | null>(null);

    useEffect(() => {
        let cancelled = false;
        (async () => {
            setEnvError(null);
            try {
                const envs = await fetchEnvironments(orgSlug, app.slug);
                if (cancelled) return;
                setEnvironments(envs);
                setSelectedEnvSlug((current) =>
                    current && envs.some((e) => e.slug === current)
                        ? current
                        : (envs[0]?.slug ?? null),
                );
            } catch (err) {
                if (!cancelled) {
                    setEnvError(
                        err instanceof Error
                            ? err.message
                            : "Failed to load environments",
                    );
                }
            }
        })();
        return () => {
            cancelled = true;
        };
    }, [orgSlug, app.id]);

    const selectedEnv =
        environments.find((e) => e.slug === selectedEnvSlug) ?? null;

    useEffect(() => {
        if (!selectedEnv) return;
        let cancelled = false;
        (async () => {
            setSecretError(null);
            try {
                const items = await fetchSecrets(orgSlug, app.slug, selectedEnv.slug);
                if (cancelled) return;
                setSecrets(items);
            } catch (err) {
                if (!cancelled) {
                    setSecretError(
                        err instanceof Error ? err.message : "Failed to load secrets",
                    );
                }
            }
        })();
        closeBatch();
        return () => {
            cancelled = true;
        };
    }, [orgSlug, app.slug, selectedEnv]);

    const resetSecrets = () => {
        setSecrets([]);
        setSelectedEnvSlug(null);
        setEditingKey(null);
        setIsSecretFormOpen(false);
        setRevealed({});
        setSecretForm({ key: "", kind: "env", value: "" });
        setSecretError(null);
    };

    const openNewSecret = () => {
        setEditingKey(null);
        setIsSecretFormOpen(true);
        setSecretForm({ key: "", kind: "env", value: "" });
        setSecretError(null);
    };

    const closeSecretForm = () => {
        setEditingKey(null);
        setIsSecretFormOpen(false);
        setSecretForm({ key: "", kind: "env", value: "" });
        setSecretError(null);
    };

    const openEditSecret = (s: WorkspaceEnvSecret) => {
        setEditingKey(s.key);
        setIsSecretFormOpen(true);
        setSecretForm({ key: s.key, kind: s.kind, value: s.value ?? "" });
        setSecretError(null);
    };

    const secretFormVisible = isSecretFormOpen || editingKey !== null;

    // ─── Environment handlers ─────────────────────────────────────

    const handleEnvNameChange = (value: string) => {
        setEnvForm((f) => ({ ...f, name: value }));
        if (!envSlugTouched) setEnvForm((f) => ({ ...f, slug: slugify(value) }));
    };

    const handleCreateEnv = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!envForm.name.trim() || !envForm.slug.trim()) return;
        setIsSavingEnv(true);
        try {
            const created = await createEnvironment(orgSlug, app.slug, {
                name: envForm.name.trim(),
                slug: slugify(envForm.slug),
                isProtected: envForm.isProtected,
            });
            addToast("Environment created", "success");
            setEnvironments((envs) => [...envs, created]);
            setSelectedEnvSlug(created.slug);
            setIsEnvFormOpen(false);
            setEnvForm({ name: "", slug: "", isProtected: false });
            setEnvSlugTouched(false);
        } catch (err) {
            addToast(
                err instanceof Error ? err.message : "Failed to create environment",
                "error",
            );
        } finally {
            setIsSavingEnv(false);
        }
    };

    const handleDeleteEnv = async (env: WorkspaceEnvironment) => {
        try {
            await deleteEnvironment(orgSlug, app.slug, env.slug);
            addToast("Environment deleted", "success");
            setEnvironments((envs) => {
                const next = envs.filter((e) => e.slug !== env.slug);
                setSelectedEnvSlug(
                    (current) =>
                        current && next.some((e) => e.slug === current)
                            ? current
                            : (next[0]?.slug ?? null),
                );
                return next;
            });
            resetSecrets();
        } catch (err) {
            addToast(
                err instanceof Error ? err.message : "Failed to delete environment",
                "error",
            );
        }
    };

    // ─── Secret handlers ──────────────────────────────────────────

    const handleSaveSecret = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!selectedEnv) return;
        if (editingKey) {
            if (!secretForm.value) return;
            setIsSavingSecret(true);
            try {
                const saved = await updateSecret(
                    orgSlug,
                    app.slug,
                    selectedEnv.slug,
                    editingKey,
                    { kind: secretForm.kind, value: secretForm.value },
                );
                addToast("Secret updated", "success");
                setSecrets((items) =>
                    items.map((s) =>
                        s.key === editingKey ? { ...s, ...saved } : s,
                    ),
                );
                closeSecretForm();
            } catch (err) {
                setSecretError(
                    err instanceof Error ? err.message : "Failed to update secret",
                );
            } finally {
                setIsSavingSecret(false);
            }
            return;
        }

        if (!secretForm.key.trim()) return;
        setIsSavingSecret(true);
        try {
            await createSecret(orgSlug, app.slug, selectedEnv.slug, {
                key: secretForm.key.trim(),
                kind: secretForm.kind,
                value: secretForm.value,
            });
            addToast("Secret saved", "success");
            const items = await fetchSecrets(orgSlug, app.slug, selectedEnv.slug);
            setSecrets(items);
            closeSecretForm();
        } catch (err) {
            setSecretError(
                err instanceof Error ? err.message : "Failed to save secret",
            );
        } finally {
            setIsSavingSecret(false);
        }
    };

    const handleDeleteSecret = async (s: WorkspaceEnvSecret) => {
        if (!selectedEnv) return;
        try {
            await deleteSecret(orgSlug, app.slug, selectedEnv.slug, s.key);
            addToast("Secret deleted", "success");
            setSecrets((items) => items.filter((x) => x.key !== s.key));
            setRevealed((r) => {
                const next = { ...r };
                delete next[s.key];
                return next;
            });
            if (editingKey === s.key) {
                closeSecretForm();
            }
        } catch (err) {
            addToast(
                err instanceof Error ? err.message : "Failed to delete secret",
                "error",
            );
        }
    };

    const toggleReveal = (key: string) => {
        setRevealed((r) => ({ ...r, [key]: !r[key] }));
    };

    // ─── Batch import / replication handlers ─────────────────────

    const existingKeys = new Set(secrets.map((s) => s.key));

    const openDraftFromEnv = async (source: WorkspaceEnvironment) => {
        setBatchError(null);
        try {
            const sourceSecrets = await fetchSecrets(
                orgSlug,
                app.slug,
                source.slug,
            );
            const rows: DraftSecretRow[] = [];
            let skipped = 0;
            for (const s of sourceSecrets) {
                if (existingKeys.has(s.key)) {
                    skipped += 1;
                    continue;
                }
                rows.push({ key: s.key, kind: s.kind, value: "" });
            }
            if (rows.length === 0) {
                addToast(
                    `All ${sourceSecrets.length} keys of "${source.name}" already exist here`,
                    "info",
                );
                setIsCopyPickerOpen(false);
                return;
            }
            setDraftRows(rows);
            setDraftSkipped(skipped);
            setDraftSourceLabel(`keys copied from ${source.name}`);
            setIsCopyPickerOpen(false);
        } catch (err) {
            setBatchError(
                err instanceof Error
                    ? err.message
                    : "Failed to read the source environment",
            );
        }
    };

    const openDraftFromFile = (text: string, fileName: string) => {
        const entries = parseDotEnv(text);
        const byKey = new Map<string, string>();
        for (const entry of entries) byKey.set(entry.key, entry.value);

        const rows: DraftSecretRow[] = [];
        let skipped = 0;
        for (const [key, value] of byKey) {
            if (existingKeys.has(key)) {
                skipped += 1;
                continue;
            }
            rows.push({ key, kind: "env", value });
        }
        if (rows.length === 0) {
            addToast("Nothing new to import — every key already exists here", "info");
            return;
        }
        setDraftRows(rows);
        setDraftSkipped(skipped);
        setDraftSourceLabel(`imported from ${fileName}`);
    };

    const handleFilePick = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;
        void file.text().then((text) => {
            openDraftFromFile(text, file.name);
        });
        e.target.value = "";
    };

    const updateDraftRow = (
        index: number,
        patch: Partial<DraftSecretRow>,
    ) => {
        setDraftRows((rows) =>
            rows.map((row, i) => (i === index ? { ...row, ...patch } : row)),
        );
    };

    const removeDraftRow = (index: number) => {
        setDraftRows((rows) => rows.filter((_, i) => i !== index));
    };

    const closeBatch = () => {
        setDraftRows([]);
        setDraftSourceLabel(null);
        setDraftSkipped(0);
        setBatchError(null);
        setIsCopyPickerOpen(false);
    };

    const handleSaveBatch = async () => {
        if (!selectedEnv) return;
        const missingKey = draftRows.some((row) => !row.key.trim());
        if (missingKey) {
            setBatchError("Every row needs a key.");
            return;
        }
        setIsBatchSaving(true);
        setBatchError(null);
        try {
            for (const row of draftRows) {
                await createSecret(orgSlug, app.slug, selectedEnv.slug, {
                    key: row.key.trim(),
                    kind: row.kind,
                    value: row.value,
                });
            }
            addToast(
                `Created ${draftRows.length} secret${draftRows.length === 1 ? "" : "s"} in ${selectedEnv.slug}`,
                "success",
            );
            setSecrets(await fetchSecrets(orgSlug, app.slug, selectedEnv.slug));
            closeBatch();
        } catch (err) {
            setBatchError(
                err instanceof Error
                    ? err.message
                    : "Failed to create batch of secrets",
            );
        } finally {
            setIsBatchSaving(false);
        }
    };

    return (
        <>
            <section className="plate p-6">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <Nameplate>Secrets</Nameplate>
                <span className="font-mono text-xs text-sand-500">
                    environments → secrets
                </span>
            </div>

            <div className="mt-5 grid grid-cols-1 gap-5 lg:grid-cols-3">
                {/* ── Environments ── */}
                <div className="flex flex-col gap-3 rounded-lg border border-brass-500/15 bg-coal-950 p-4 lg:col-span-1">
                    <div className="flex items-center justify-between gap-2">
                        <h3 className="text-xs font-semibold uppercase tracking-[0.2em] text-sand-500">
                            Environments
                        </h3>
                        <button
                            type="button"
                            onClick={() => setIsEnvFormOpen((v) => !v)}
                            className="lk-btn lk-btn--ghost !px-2.5 !py-1 text-xs"
                        >
                            {isEnvFormOpen ? "Close" : "+ New"}
                        </button>
                    </div>

                    {isEnvFormOpen && (
                        <form
                            onSubmit={handleCreateEnv}
                            className="flex flex-col gap-2.5 border-b border-brass-500/10 pb-3"
                        >
                            <Field label="Name *">
                                <input
                                    type="text"
                                    value={envForm.name}
                                    onChange={(e) =>
                                        handleEnvNameChange(e.target.value)
                                    }
                                    placeholder="e.g., Production"
                                    className="lk-input"
                                />
                            </Field>
                            <Field label="Slug *">
                                <input
                                    type="text"
                                    value={envForm.slug}
                                    onChange={(e) => {
                                        setEnvSlugTouched(true);
                                        setEnvForm((f) => ({
                                            ...f,
                                            slug: slugify(e.target.value),
                                        }));
                                    }}
                                    placeholder="prod"
                                    className="lk-input font-mono"
                                />
                            </Field>
                            <label className="flex cursor-pointer items-center gap-2 text-sm text-sand-400">
                                <input
                                    type="checkbox"
                                    checked={envForm.isProtected}
                                    onChange={(e) =>
                                        setEnvForm((f) => ({
                                            ...f,
                                            isProtected: e.target.checked,
                                        }))
                                    }
                                    className="h-4 w-4 accent-brass-500"
                                />
                                Protected
                            </label>
                            <button
                                type="submit"
                                disabled={isSavingEnv}
                                className="lk-btn lk-btn--primary"
                            >
                                {isSavingEnv ? "Creating..." : "Create Environment"}
                            </button>
                        </form>
                    )}

                    <div className="flex flex-col gap-2">
                        {envError ? (
                            <p className="rounded-lg border border-vermillion-500/25 bg-vermillion-500/10 px-3 py-2 text-xs text-vermillion-300">
                                {envError}
                            </p>
                        ) : environments.length === 0 ? (
                            <p className="rounded-lg border border-dashed border-brass-500/20 px-3 py-5 text-center text-xs text-sand-600">
                                No environments yet. Each environment gets its own
                                data key.
                            </p>
                        ) : (
                            environments.map((env) => (
                                <div
                                    key={env.id}
                                    className={`flex items-center justify-between gap-2 rounded-lg border px-3 py-2.5 transition-colors ${
                                        selectedEnv?.id === env.id
                                            ? "border-brass-400/40 bg-brass-500/10"
                                            : "border-brass-500/15 bg-coal-900 hover:bg-coal-850"
                                    }`}
                                >
                                    <button
                                        type="button"
                                        onClick={() => setSelectedEnvSlug(env.slug)}
                                        className="min-w-0 flex-1 text-left"
                                    >
                                        <p className="truncate text-sm font-medium text-ivory">
                                            {env.name}
                                        </p>
                                        <p className="truncate font-mono text-xs text-sand-500">
                                            {env.slug}
                                            {env.isProtected && (
                                                <span className="ml-2 text-patina-400">
                                                    · protected
                                                </span>
                                            )}
                                        </p>
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => handleDeleteEnv(env)}
                                        title={`Delete ${env.name}`}
                                        className="shrink-0 text-xs text-sand-600 transition-colors hover:text-vermillion-300"
                                    >
                                        Delete
                                    </button>
                                </div>
                            ))
                        )}
                    </div>
                </div>

                {/* ── Secrets ── */}
                <div className="flex flex-col gap-3 rounded-lg border border-brass-500/15 bg-coal-950 p-4 lg:col-span-2">
                    <div className="flex items-center justify-between gap-2">
                        <h3 className="text-xs font-semibold uppercase tracking-[0.2em] text-sand-500">
                            Secrets
                            {selectedEnv && (
                                <span className="ml-2 font-mono normal-case tracking-normal text-sand-600">
                                    · {selectedEnv.slug}
                                </span>
                            )}
                        </h3>
                        <div className="flex items-center gap-2">
                            <button
                                type="button"
                                onClick={() => setIsCopyPickerOpen((v) => !v)}
                                disabled={
                                    !selectedEnv ||
                                    environments.length < 2 ||
                                    draftRows.length > 0
                                }
                                title="Replicate keys from another environment"
                                className="lk-btn lk-btn--ghost !px-2.5 !py-1 text-xs disabled:cursor-not-allowed disabled:opacity-40"
                            >
                                Replicate
                            </button>
                            <button
                                type="button"
                                onClick={() => fileInputRef.current?.click()}
                                disabled={!selectedEnv || draftRows.length > 0}
                                title="Import secrets from a .env file"
                                className="lk-btn lk-btn--ghost !px-2.5 !py-1 text-xs disabled:cursor-not-allowed disabled:opacity-40"
                            >
                                Import .env
                            </button>
                            <button
                                type="button"
                                onClick={secretFormVisible ? closeSecretForm : openNewSecret}
                                disabled={!selectedEnv || draftRows.length > 0}
                                className="lk-btn lk-btn--ghost !px-2.5 !py-1 text-xs disabled:cursor-not-allowed disabled:opacity-40"
                            >
                                {secretFormVisible ? "Close" : "+ Add Secret"}
                            </button>
                        </div>
                    </div>

                    <input
                        ref={fileInputRef}
                        type="file"
                        accept=".env,text/plain"
                        className="hidden"
                        onChange={handleFilePick}
                    />

                    {isCopyPickerOpen && selectedEnv && (
                        <div className="flex flex-col gap-2 rounded-lg border border-brass-500/15 bg-coal-950 p-3">
                            <span className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                                Copy keys from
                            </span>
                            <div className="flex flex-wrap gap-2">
                                {environments
                                    .filter((e) => e.slug !== selectedEnv.slug)
                                    .map((env) => (
                                        <button
                                            key={env.id}
                                            type="button"
                                            onClick={() =>
                                                void openDraftFromEnv(env)
                                            }
                                            className="lk-btn lk-btn--ghost !px-2.5 !py-1 text-xs"
                                        >
                                            {env.name}
                                        </button>
                                    ))}
                            </div>
                            <button
                                type="button"
                                onClick={() => setIsCopyPickerOpen(false)}
                                className="self-start text-xs text-sand-600 transition-colors hover:text-sand-400"
                            >
                                Close
                            </button>
                        </div>
                    )}

                    {!selectedEnv ? (
                        <p className="rounded-lg border border-dashed border-brass-500/20 px-3 py-5 text-center text-xs text-sand-600">
                            Select an environment to manage its secrets.
                        </p>
                    ) : draftRows.length > 0 ? (
                        <div className="flex flex-col gap-3 border-b border-brass-500/10 pb-3">
                            <div>
                                <Nameplate tone="muted">
                                    Batch — {draftSourceLabel ?? "secrets"}
                                </Nameplate>
                                <p className="mt-1.5 text-xs leading-relaxed text-sand-600">
                                    {draftRows.length} new{" "}
                                    {draftRows.length === 1
                                        ? "secret"
                                        : "secrets"}{" "}
                                    to create
                                    {draftSkipped > 0
                                        ? ` · ${draftSkipped} skipped (already in ${selectedEnv.slug})`
                                        : ""}
                                    . Only keys were imported — values stay
                                    unique to this environment.
                                </p>
                            </div>

                            <div className="flex flex-col gap-2">
                                {draftRows.map((row, index) => (
                                    <div
                                        key={index}
                                        className="grid grid-cols-[5.5rem_12rem_1fr_auto] items-start gap-2 rounded-md border border-brass-500/15 bg-coal-900 p-2.5"
                                    >
                                        <select
                                            value={row.kind}
                                            onChange={(e) =>
                                                updateDraftRow(index, {
                                                    kind: e.target
                                                        .value as SharedSecretKind,
                                                })
                                            }
                                            className="lk-input !py-1 text-xs"
                                        >
                                            <option value="env">env</option>
                                            <option value="note">note</option>
                                            <option value="file">file</option>
                                            <option value="json">json</option>
                                        </select>
                                        <input
                                            type="text"
                                            value={row.key}
                                            onChange={(e) =>
                                                updateDraftRow(index, {
                                                    key: e.target.value,
                                                })
                                            }
                                            placeholder="KEY"
                                            className="lk-input !py-1 font-mono text-xs"
                                        />
                                        <input
                                            type="text"
                                            value={row.value}
                                            onChange={(e) =>
                                                updateDraftRow(index, {
                                                    value: e.target.value,
                                                })
                                            }
                                            placeholder="value (fill per environment)"
                                            className="lk-input !py-1 font-mono text-xs"
                                        />
                                        <button
                                            type="button"
                                            onClick={() =>
                                                removeDraftRow(index)
                                            }
                                            className="shrink-0 text-xs text-sand-600 transition-colors hover:text-vermillion-300"
                                        >
                                            Remove
                                        </button>
                                    </div>
                                ))}
                            </div>

                            <button
                                type="button"
                                onClick={() =>
                                    setDraftRows((rows) => [
                                        ...rows,
                                        { key: "", kind: "env", value: "" },
                                    ])
                                }
                                className="lk-btn lk-btn--ghost !px-2.5 !py-1 self-start text-xs"
                            >
                                + Add Row
                            </button>

                            {batchError && (
                                <p className="rounded-lg border border-vermillion-500/25 bg-vermillion-500/10 px-3 py-2 text-xs text-vermillion-300">
                                    {batchError}
                                </p>
                            )}

                            <div className="flex items-center gap-2">
                                <button
                                    type="button"
                                    onClick={() => void handleSaveBatch()}
                                    disabled={isBatchSaving}
                                    className="lk-btn lk-btn--primary"
                                >
                                    {isBatchSaving
                                        ? "Creating…"
                                        : `Create ${draftRows.length}`}
                                </button>
                                <button
                                    type="button"
                                    onClick={closeBatch}
                                    disabled={isBatchSaving}
                                    className="lk-btn lk-btn--ghost"
                                >
                                    Cancel
                                </button>
                            </div>
                        </div>
                    ) : secretFormVisible ? (
                        <form
                            onSubmit={handleSaveSecret}
                            className="flex flex-col gap-2.5 border-b border-brass-500/10 pb-3"
                        >
                            <div>
                                <Nameplate tone="muted">
                                    {editingKey ? "Edit Secret" : "New Secret"}
                                </Nameplate>
                            </div>
                            <Field label={editingKey ? "Key (read-only)" : "Key *"}>
                                <input
                                    type="text"
                                    value={secretForm.key}
                                    disabled={Boolean(editingKey)}
                                    onChange={(e) =>
                                        setSecretForm((f) => ({
                                            ...f,
                                            key: e.target.value,
                                        }))
                                    }
                                    placeholder="DATABASE_URL"
                                    className="lk-input font-mono"
                                />
                            </Field>
                            <Field label="Kind *">
                                <select
                                    value={secretForm.kind}
                                    onChange={(e) =>
                                        setSecretForm((f) => ({
                                            ...f,
                                            kind: e.target.value as SharedSecretKind,
                                        }))
                                    }
                                    className="lk-input"
                                >
                                    <option value="env">env</option>
                                    <option value="note">note</option>
                                    <option value="file">file</option>
                                    <option value="json">json</option>
                                </select>
                            </Field>
                            <Field label="Value *">
                                <textarea
                                    value={secretForm.value}
                                    onChange={(e) =>
                                        setSecretForm((f) => ({
                                            ...f,
                                            value: e.target.value,
                                        }))
                                    }
                                    rows={3}
                                    placeholder="postgres://user:pass@db:5432/app"
                                    className="lk-input resize-y font-mono"
                                />
                            </Field>
                            {secretError && (
                                <p className="rounded-lg border border-vermillion-500/25 bg-vermillion-500/10 px-3 py-2 text-xs text-vermillion-300">
                                    {secretError}
                                </p>
                            )}
                            <button
                                type="submit"
                                disabled={isSavingSecret}
                                className="lk-btn lk-btn--primary"
                            >
                                {isSavingSecret
                                    ? "Saving..."
                                    : editingKey
                                      ? "Save Changes"
                                      : "Add Secret"}
                            </button>
                        </form>
                    ) : null}

                    {draftRows.length === 0 && (
                        <div className="flex flex-col gap-2">
                            {secretError && !secretFormVisible && (
                                <p className="rounded-lg border border-vermillion-500/25 bg-vermillion-500/10 px-3 py-2 text-xs text-vermillion-300">
                                    {secretError}
                                </p>
                            )}
                            {secrets.length === 0 ? (
                                <p className="rounded-lg border border-dashed border-brass-500/20 px-3 py-5 text-center text-xs text-sand-600">
                                    {secretFormVisible
                                        ? "Fill in the form above to add the first secret."
                                        : "No secrets stored in this environment yet."}
                                </p>
                            ) : (
                                secrets.map((s) => (
                                    <div
                                        key={s.key}
                                        className="rounded-lg border border-brass-500/15 bg-coal-900 px-3 py-2.5"
                                    >
                                        <div className="flex items-center justify-between gap-2">
                                            <div className="flex min-w-0 items-center gap-2">
                                                <p className="truncate font-mono text-sm font-medium text-ivory">
                                                    {s.key}
                                                </p>
                                                <KindBadge kind={s.kind} />
                                                <span className="shrink-0 font-mono text-[0.625rem] text-sand-600">
                                                    v{s.version}
                                                </span>
                                            </div>
                                            <div className="flex shrink-0 items-center gap-2 text-xs">
                                                <button
                                                    type="button"
                                                    onClick={() =>
                                                        toggleReveal(s.key)
                                                    }
                                                    className="text-sand-500 transition-colors hover:text-brass-300"
                                                >
                                                    {revealed[s.key]
                                                        ? "Hide"
                                                        : "Show"}
                                                </button>
                                                <button
                                                    type="button"
                                                    onClick={() =>
                                                        openEditSecret(s)
                                                    }
                                                    className="text-sand-500 transition-colors hover:text-brass-300"
                                                >
                                                    Edit
                                                </button>
                                                <button
                                                    type="button"
                                                    onClick={() =>
                                                        handleDeleteSecret(s)
                                                    }
                                                    className="text-sand-600 transition-colors hover:text-vermillion-300"
                                                >
                                                    Delete
                                                </button>
                                            </div>
                                        </div>
                                        <pre className="mt-2 overflow-x-auto whitespace-pre-wrap break-all rounded-md border border-brass-500/10 bg-coal-950 px-3 py-2 font-mono text-xs text-patina-300">
                                            {revealed[s.key]
                                                ? s.value || "(empty)"
                                                : "••••••••••••••••"}
                                        </pre>
                                        <p className="mt-1.5 font-mono text-[0.625rem] text-sand-600">
                                            updated{" "}
                                            {new Date(
                                                s.updatedAt,
                                            ).toLocaleString()}
                                        </p>
                                    </div>
                                ))
                            )}
                        </div>
                    )}
                </div>
            </div>

            <p className="mt-4 text-xs leading-relaxed text-sand-600">
                Secrets are encrypted per environment; the server unwraps the
                environment key when you view them here. Machines fetch the same
                values through an environment-scoped API key.
            </p>
        </section>

        {selectedEnv && (
            <div className="mt-5">
                <MachineAccess orgSlug={orgSlug} app={app} env={selectedEnv} />
            </div>
        )}
        </>
    );
}