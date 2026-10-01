import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import type { CreateApplicationInput, WorkspaceApplication } from "~/types";
import {
    createApplication,
    deleteApplication,
    fetchApplications,
} from "~/lib/api/sharedSecrets";
import { useToast } from "~/providers/toast";
import Nameplate from "./Nameplate";

const slugify = (value: string) =>
    value
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "");

export default function WorkspaceApplications({ slug }: { slug: string }) {
    const { addToast } = useToast();
    const navigate = useNavigate();

    const [applications, setApplications] = useState<WorkspaceApplication[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [loadError, setLoadError] = useState<string | null>(null);

    const [isAppFormOpen, setIsAppFormOpen] = useState(false);
    const [appForm, setAppForm] = useState<CreateApplicationInput>({
        name: "",
        slug: "",
    });
    const [appSlugTouched, setAppSlugTouched] = useState(false);
    const [isSaving, setIsSaving] = useState(false);

    const [deletingId, setDeletingId] = useState<string | null>(null);

    useEffect(() => {
        let cancelled = false;
        (async () => {
            setIsLoading(true);
            setLoadError(null);
            try {
                const apps = await fetchApplications(slug);
                if (!cancelled) setApplications(apps);
            } catch (err) {
                if (!cancelled) {
                    setLoadError(
                        err instanceof Error
                            ? err.message
                            : "Failed to load applications",
                    );
                }
            } finally {
                if (!cancelled) setIsLoading(false);
            }
        })();
        return () => {
            cancelled = true;
        };
    }, [slug]);

    const handleNameChange = (value: string) => {
        setAppForm((f) => ({ ...f, name: value }));
        if (!appSlugTouched) setAppForm((f) => ({ ...f, slug: slugify(value) }));
    };

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!appForm.name.trim() || !appForm.slug.trim()) return;
        setIsSaving(true);
        try {
            const created = await createApplication(slug, {
                name: appForm.name.trim(),
                slug: slugify(appForm.slug),
            });
            addToast("Application created", "success");
            navigate(`/workspaces/${slug}/applications/${created.slug}`);
        } catch (err) {
            addToast(
                err instanceof Error ? err.message : "Failed to create application",
                "error",
            );
        } finally {
            setIsSaving(false);
        }
    };

    const handleDelete = async (app: WorkspaceApplication) => {
        if (deletingId) return;
        setDeletingId(app.id);
        try {
            await deleteApplication(slug, app.slug);
            addToast("Application deleted", "success");
            setApplications((apps) => apps.filter((a) => a.id !== app.id));
        } catch (err) {
            addToast(
                err instanceof Error ? err.message : "Failed to delete application",
                "error",
            );
        } finally {
            setDeletingId(null);
        }
    };

    return (
        <section className="plate p-6">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <Nameplate>Applications</Nameplate>
                <span className="font-mono text-xs text-sand-500">
                    {applications.length} on file · each holds its own environments
                </span>
            </div>

            <form
                onSubmit={handleCreate}
                className="mt-4 flex flex-wrap items-end gap-3"
            >
                <div className="flex min-w-[180px] flex-1 flex-col gap-1.5">
                    <label className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                        Application Name
                    </label>
                    <input
                        type="text"
                        value={appForm.name}
                        onChange={(e) => handleNameChange(e.target.value)}
                        placeholder="e.g., Website"
                        className="lk-input"
                    />
                </div>
                <div className="flex min-w-[140px] flex-1 flex-col gap-1.5">
                    <label className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                        Slug
                    </label>
                    <input
                        type="text"
                        value={appForm.slug}
                        onChange={(e) => {
                            setAppSlugTouched(true);
                            setAppForm((f) => ({ ...f, slug: slugify(e.target.value) }));
                        }}
                        placeholder="website"
                        className="lk-input font-mono"
                    />
                </div>
                <button
                    type="submit"
                    disabled={isSaving}
                    className="lk-btn lk-btn--primary"
                >
                    {isSaving ? "Creating..." : "Create Application"}
                </button>
            </form>

            <div className="mt-5 flex flex-col gap-2">
                {loadError ? (
                    <p className="rounded-lg border border-vermillion-500/25 bg-vermillion-500/10 px-4 py-3 text-sm text-vermillion-300">
                        {loadError}
                    </p>
                ) : isLoading ? (
                    <p className="rounded-lg border border-dashed border-brass-500/20 px-4 py-6 text-center text-sm text-sand-600">
                        Loading applications...
                    </p>
                ) : applications.length === 0 ? (
                    <p className="rounded-lg border border-dashed border-brass-500/20 px-4 py-6 text-center text-sm text-sand-600">
                        No applications yet. Create one above to start adding
                        environments and secrets.
                    </p>
                ) : (
                    applications.map((app) => (
                        <div
                            key={app.id}
                            className="group flex items-center justify-between gap-4 rounded-lg border border-brass-500/15 bg-coal-950 px-4 py-3"
                        >
                            <button
                                type="button"
                                onClick={() =>
                                    navigate(
                                        `/workspaces/${slug}/applications/${app.slug}`,
                                    )
                                }
                                className="min-w-0 flex-1 text-left"
                            >
                                <p className="truncate text-sm font-medium text-ivory">
                                    {app.name}
                                </p>
                                <p className="truncate font-mono text-xs text-sand-500">
                                    {app.slug}
                                    {app.description && (
                                        <span className="ml-2 text-sand-600">
                                            · {app.description}
                                        </span>
                                    )}
                                </p>
                            </button>
                            <div className="flex shrink-0 items-center gap-3">
                                <span className="font-mono text-[0.625rem] text-sand-600">
                                    {new Date(app.createdAt).toLocaleDateString()}
                                </span>
                                <button
                                    type="button"
                                    onClick={() =>
                                        navigate(
                                            `/workspaces/${slug}/applications/${app.slug}`,
                                        )
                                    }
                                    className="lk-btn lk-btn--ghost !px-2.5 !py-1 text-xs"
                                >
                                    Manage
                                </button>
                                <button
                                    type="button"
                                    onClick={() => handleDelete(app)}
                                    disabled={deletingId === app.id}
                                    title={`Delete ${app.name}`}
                                    className="text-xs text-sand-600 transition-colors hover:text-vermillion-300 disabled:opacity-40"
                                >
                                    Delete
                                </button>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </section>
    );
}