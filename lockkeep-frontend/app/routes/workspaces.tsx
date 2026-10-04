import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router";
import { requireAuth } from "~/lib/auth-guard";
import type { Route } from "./+types/workspaces";
import {
  createWorkspace,
  fetchWorkspaces,
} from "~/lib/api/workspaces";
import type { Workspace } from "~/types";
import { useToast } from "~/providers/toast";
import Nameplate from "~/components/Nameplate";
import VaultDial from "~/components/VaultDial";

export const clientLoader = () => {
  return requireAuth();
};

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Workspaces — LockKeep" },
    { name: "description", content: "Manage your LockKeep workspaces: shared spaces for teams, environments, and secrets." },
    { name: "robots", content: "noindex, nofollow" },
  ];
}

const slugify = (value: string) =>
  value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");

export default function Workspaces() {
  const { addToast } = useToast();
  const navigate = useNavigate();

  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugTouched, setSlugTouched] = useState(false);
  const [isCreating, setIsCreating] = useState(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const ws = await fetchWorkspaces();
        if (!cancelled) setWorkspaces(ws);
      } catch (err) {
        if (!cancelled) {
          addToast(
            err instanceof Error ? err.message : "Failed to load workspaces",
            "error",
          );
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [addToast]);

  const openCreate = () => {
    setName("");
    setSlug("");
    setSlugTouched(false);
    setIsCreateOpen(true);
  };

  const handleNameChange = (value: string) => {
    setName(value);
    if (!slugTouched) setSlug(slugify(value));
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !slug.trim()) {
      addToast("Name and slug are required", "error");
      return;
    }

    setIsCreating(true);
    try {
      const ws = await createWorkspace({ name: name.trim(), slug: slugify(slug) });
      addToast("Workspace established", "success");
      navigate(`/workspaces/${ws.slug}`);
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to create workspace",
        "error",
      );
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div>
      <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <Nameplate>Shared Floor</Nameplate>
          <h1 className="mt-2 font-display text-3xl font-semibold tracking-tight text-ivory">
            Your Workspaces
          </h1>
        </div>
        <button onClick={openCreate} className="lk-btn lk-btn--primary">
          <svg
            width="15"
            height="15"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.2"
            strokeLinecap="round"
          >
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          Establish Workspace
        </button>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center gap-3 py-24 text-sand-500">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-brass-600 border-t-brass-400" />
          <span>Registering workspaces...</span>
        </div>
      ) : workspaces.length === 0 ? (
        <div className="flex flex-col items-center py-20 text-center">
          <VaultDial size={120} className="text-brass-700 opacity-60" />
          <p className="mt-5 font-display text-lg font-medium text-sand-400">
            No workspaces on file
          </p>
          <p className="mt-1 mb-6 max-w-sm text-sm text-sand-600">
            Establish a workspace to share secrets, teams, and environments with
            your crew.
          </p>
          <button onClick={openCreate} className="lk-btn lk-btn--primary">
            Establish Your First Workspace
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
          {workspaces.map((ws) => (
            <Link
              key={ws.id}
              to={`/workspaces/${ws.slug}`}
              className="plate group p-6 transition-colors hover:border-brass-400/40"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <h2 className="truncate font-display text-xl font-medium text-ivory">
                    {ws.name}
                  </h2>
                  <p className="mt-1 truncate font-mono text-sm text-brass-400/80">
                    /workspaces/{ws.slug}
                  </p>
                </div>
                <span className="mt-1 h-2 w-2 shrink-0 rounded-full bg-patina-400" />
              </div>

              <div className="mt-5 flex items-center gap-2 border-t border-brass-500/10 pt-3">
                <span className="text-[0.625rem] uppercase tracking-[0.18em] text-sand-600">
                  Registered
                </span>
                <span className="h-px flex-1 bg-brass-500/15" />
                <span className="font-mono text-[0.625rem] text-sand-600">
                  {new Date(ws.createdAt).toLocaleDateString()}
                </span>
              </div>

              <div className="mt-4 flex items-center gap-2 text-sm font-medium text-brass-300">
                Enter workspace
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  className="transition-transform group-hover:translate-x-1"
                >
                  <polyline points="9 18 15 12 9 6" />
                </svg>
              </div>
            </Link>
          ))}
        </div>
      )}

      {/* Create modal */}
      {isCreateOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-coal-950/85 p-4 backdrop-blur-sm">
          <div
            className="plate w-full max-w-md p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="mb-5">
              <Nameplate>New Workspace</Nameplate>
              <h2 className="mt-1.5 font-display text-xl font-medium text-ivory">
                Establish a Workspace
              </h2>
            </div>

            <form onSubmit={handleCreate} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
                  Name *
                </label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => handleNameChange(e.target.value)}
                  placeholder="e.g., Acme Payments"
                  required
                  autoFocus
                  className="lk-input"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
                  Slug *
                </label>
                <div className="relative">
                  <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 font-mono text-sm text-sand-600">
                    /workspaces/
                  </span>
                  <input
                    type="text"
                    value={slug}
                    onChange={(e) => {
                      setSlugTouched(true);
                      setSlug(slugify(e.target.value));
                    }}
                    placeholder="acme-payments"
                    required
                    className="lk-input pl-28 font-mono"
                  />
                </div>
              </div>

              <div className="mt-2 flex justify-end gap-3 border-t border-brass-500/15 pt-4">
                <button
                  type="button"
                  onClick={() => setIsCreateOpen(false)}
                  className="lk-btn lk-btn--ghost"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isCreating}
                  className="lk-btn lk-btn--primary"
                >
                  {isCreating ? "Establishing..." : "Establish"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}