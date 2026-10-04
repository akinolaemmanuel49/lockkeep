import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { requireAuth } from "~/lib/auth-guard";
import type { Route } from "./+types/workspace.$slug";
import {
  createWorkspaceTeam,
  deleteWorkspace,
  fetchWorkspace,
  fetchWorkspaceTeams,
  updateWorkspace,
} from "~/lib/api/workspaces";
import type { Workspace, WorkspaceTeam } from "~/types";
import { useToast } from "~/providers/toast";
import { useAuth } from "~/providers/auth";
import Nameplate from "~/components/Nameplate";
import WorkspaceApplications from "~/components/WorkspaceApplications";

export const clientLoader = () => {
  return requireAuth();
};

export function meta({}: Route.MetaArgs) {
  return [{ title: "Workspace — LockKeep" }];
}

const slugify = (value: string) =>
  value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");

function DetailRow({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-brass-500/10 py-2.5 last:border-0">
      <span className="text-[0.625rem] font-semibold uppercase tracking-[0.22em] text-sand-600">
        {label}
      </span>
      <span className="text-right font-mono text-sm text-sand-300">
        {children}
      </span>
    </div>
  );
}

export default function WorkspaceDetail() {
  const { slug } = useParams();
  const navigate = useNavigate();
  const { addToast } = useToast();
  const { user } = useAuth();

  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [teams, setTeams] = useState<WorkspaceTeam[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const [teamName, setTeamName] = useState("");
  const [teamSlug, setTeamSlug] = useState("");
  const [teamSlugTouched, setTeamSlugTouched] = useState(false);
  const [isCreatingTeam, setIsCreatingTeam] = useState(false);

  const [isEditOpen, setIsEditOpen] = useState(false);
  const [editName, setEditName] = useState("");
  const [editSlug, setEditSlug] = useState("");
  const [editSlugTouched, setEditSlugTouched] = useState(false);
  const [isSavingEdit, setIsSavingEdit] = useState(false);

  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    if (!slug) return;
    let cancelled = false;

    (async () => {
      setIsLoading(true);
      setLoadError(null);
      try {
        const [ws, ts] = await Promise.all([
          fetchWorkspace(slug),
          fetchWorkspaceTeams(slug),
        ]);
        if (cancelled) return;
        setWorkspace(ws);
        setTeams(ts);
      } catch (err) {
        if (!cancelled) {
          setLoadError(
            err instanceof Error ? err.message : "Failed to load workspace",
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

  const isOwner = workspace ? user?.id === workspace.ownerId : false;

  const handleTeamNameChange = (value: string) => {
    setTeamName(value);
    if (!teamSlugTouched) setTeamSlug(slugify(value));
  };

  const handleCreateTeam = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!slug || !teamName.trim() || !teamSlug.trim()) return;

    setIsCreatingTeam(true);
    try {
      await createWorkspaceTeam(slug, {
        name: teamName.trim(),
        slug: slugify(teamSlug),
      });
      addToast("Team created", "success");
      setTeamName("");
      setTeamSlug("");
      setTeamSlugTouched(false);
      setTeams(await fetchWorkspaceTeams(slug));
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to create team",
        "error",
      );
    } finally {
      setIsCreatingTeam(false);
    }
  };

  const openEdit = () => {
    if (!workspace) return;
    setEditName(workspace.name);
    setEditSlug(workspace.slug);
    setEditSlugTouched(true);
    setIsEditOpen(true);
  };

  const handleEditNameChange = (value: string) => {
    setEditName(value);
    if (!editSlugTouched) setEditSlug(slugify(value));
  };

  const handleSaveEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!slug || !workspace) return;
    if (!editName.trim() || !editSlug.trim()) {
      addToast("Name and slug are required", "error");
      return;
    }

    setIsSavingEdit(true);
    try {
      const updated = await updateWorkspace(slug, {
        name: editName.trim(),
        slug: slugify(editSlug),
      });
      addToast("Workspace updated", "success");
      setIsEditOpen(false);
      setWorkspace(updated);
      if (updated.slug !== slug) {
        navigate(`/workspaces/${updated.slug}`, { replace: true });
      }
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to update workspace",
        "error",
      );
    } finally {
      setIsSavingEdit(false);
    }
  };

  const handleDelete = async () => {
    if (!slug) return;

    setIsDeleting(true);
    try {
      await deleteWorkspace(slug);
      addToast("Workspace deleted", "success");
      navigate("/workspaces");
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to delete workspace",
        "error",
      );
      setIsDeleteOpen(false);
    } finally {
      setIsDeleting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center gap-3 py-24 text-sand-500">
        <div className="h-5 w-5 animate-spin rounded-full border-2 border-brass-600 border-t-brass-400" />
        <span>Opening the workspace...</span>
      </div>
    );
  }

  if (loadError || !workspace) {
    return (
      <div className="mx-auto max-w-lg py-16">
        <div className="plate p-8 text-center">
          <p className="text-sm text-vermillion-300">
            {loadError ?? "Workspace not found"}
          </p>
          <Link to="/workspaces" className="lk-btn lk-btn--ghost mt-6">
            ← All Workspaces
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl">
      <Link
        to="/workspaces"
        className="mb-6 inline-flex items-center gap-2 text-sm font-medium text-sand-500 transition-colors hover:text-brass-300"
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <polyline points="15 18 9 12 15 6" />
        </svg>
        All Workspaces
      </Link>

      <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <Nameplate>Workspace</Nameplate>
          <h1 className="mt-2 font-display text-3xl font-semibold tracking-tight text-ivory">
            {workspace.name}
          </h1>
          <p className="mt-1 font-mono text-sm text-brass-400/80">
            /workspaces/{workspace.slug}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.2em]">
            <span className="h-2 w-2 rounded-full bg-patina-400" />
            Active
          </span>
          {isOwner && (
            <>
              <button onClick={openEdit} className="lk-btn lk-btn--ghost">
                Edit
              </button>
              <button
                onClick={() => setIsDeleteOpen(true)}
                className="lk-btn lk-btn--danger-ghost"
              >
                Delete
              </button>
            </>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 gap-5 lg:grid-cols-3">
        {/* Overview */}
        <section className="plate lg:col-span-1 p-6">
          <Nameplate>Overview</Nameplate>
          <div className="mt-4">
            <DetailRow label="Slug">{workspace.slug}</DetailRow>
            <DetailRow label="Owner">{workspace.ownerId.slice(0, 10)}</DetailRow>
            <DetailRow label="2FA">
              {workspace.settings.require2FA ? "Enforced" : "Optional"}
            </DetailRow>
            <DetailRow label="Default Role">
              {workspace.settings.defaultTeamRole}
            </DetailRow>
            <DetailRow label="Created">
              {new Date(workspace.createdAt).toLocaleDateString()}
            </DetailRow>
          </div>
        </section>

        <div className="flex flex-col gap-5 lg:col-span-2">
          {/* Teams */}
          <section className="plate p-6">
            <div className="flex items-center justify-between">
              <Nameplate>Teams</Nameplate>
              <span className="font-mono text-xs text-sand-500">
                {teams.length} on file
              </span>
            </div>

            <form
              onSubmit={handleCreateTeam}
              className="mt-4 flex flex-wrap items-end gap-3"
            >
              <div className="flex min-w-[180px] flex-1 flex-col gap-1.5">
                <label className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                  Team Name
                </label>
                <input
                  type="text"
                  value={teamName}
                  onChange={(e) => handleTeamNameChange(e.target.value)}
                  placeholder="e.g., Engineering"
                  className="lk-input"
                />
              </div>
              <div className="flex min-w-[140px] flex-1 flex-col gap-1.5">
                <label className="text-[0.625rem] font-semibold uppercase tracking-[0.18em] text-sand-600">
                  Slug
                </label>
                <input
                  type="text"
                  value={teamSlug}
                  onChange={(e) => {
                    setTeamSlugTouched(true);
                    setTeamSlug(slugify(e.target.value));
                  }}
                  placeholder="engineering"
                  className="lk-input font-mono"
                />
              </div>
              <button
                type="submit"
                disabled={isCreatingTeam}
                className="lk-btn lk-btn--primary"
              >
                {isCreatingTeam ? "Adding..." : "Add Team"}
              </button>
            </form>

            <div className="mt-5 flex flex-col gap-2">
              {teams.length === 0 ? (
                <p className="rounded-lg border border-dashed border-brass-500/20 px-4 py-6 text-center text-sm text-sand-600">
                  No teams established yet. Add your first team above.
                </p>
              ) : (
                teams.map((team) => (
                  <div
                    key={team.id}
                    className="flex items-center justify-between gap-4 rounded-lg border border-brass-500/15 bg-coal-950 px-4 py-3"
                  >
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium text-ivory">
                        {team.name}
                      </p>
                      <p className="truncate font-mono text-xs text-sand-500">
                        {team.slug}
                      </p>
                    </div>
                    <span className="shrink-0 font-mono text-[0.625rem] text-sand-600">
                      {new Date(team.createdAt).toLocaleDateString()}
                    </span>
                  </div>
                ))
              )}
            </div>
          </section>

          {/* Members & Roles */}
          <section className="plate p-6">
            <div className="flex items-center justify-between">
              <Nameplate>Members & Roles</Nameplate>
              <span className="font-mono text-xs text-sand-500">1 member</span>
            </div>

            <div className="mt-4 flex flex-col gap-2">
              <div className="flex items-center justify-between gap-4 rounded-lg border border-brass-500/15 bg-coal-950 px-4 py-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-ivory">
                    Workspace owner
                  </p>
                  <p className="truncate font-mono text-xs text-sand-500">
                    {workspace.ownerId}
                  </p>
                </div>
                <span className="shrink-0 rounded-full border border-brass-500/25 bg-brass-500/10 px-3 py-1 font-mono text-[0.625rem] uppercase tracking-[0.16em] text-brass-300">
                  Owner · org:owner
                </span>
              </div>
            </div>

            <p className="mt-4 text-sm leading-relaxed text-sand-500">
              Access to applications, environments, and secrets is enforced
              server-side by the workspace RBAC layer. Managing invitations and
              role assignments ships with a later phase.
            </p>
          </section>
        </div>
      </div>

      {/* Applications */}
      {slug && (
        <div className="mt-5">
          <WorkspaceApplications slug={slug} />
        </div>
      )}

      {/* Edit modal */}
      {isEditOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-coal-950/85 p-4 backdrop-blur-sm">
          <div
            className="plate w-full max-w-md p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="mb-5">
              <Nameplate>Workspace Settings</Nameplate>
              <h2 className="mt-1.5 font-display text-xl font-medium text-ivory">
                Edit Workspace
              </h2>
            </div>

            <form onSubmit={handleSaveEdit} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
                  Name *
                </label>
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => handleEditNameChange(e.target.value)}
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
                    value={editSlug}
                    onChange={(e) => {
                      setEditSlugTouched(true);
                      setEditSlug(slugify(e.target.value));
                    }}
                    required
                    className="lk-input pl-28 font-mono"
                  />
                </div>
              </div>

              <div className="mt-2 flex justify-end gap-3 border-t border-brass-500/15 pt-4">
                <button
                  type="button"
                  onClick={() => setIsEditOpen(false)}
                  className="lk-btn lk-btn--ghost"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSavingEdit}
                  className="lk-btn lk-btn--primary"
                >
                  {isSavingEdit ? "Saving..." : "Save Changes"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete confirmation */}
      {isDeleteOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-coal-950/85 p-4 backdrop-blur-sm">
          <div
            className="plate w-full max-w-md p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="mb-5">
              <Nameplate tone="danger">Destructive Action</Nameplate>
              <h2 className="mt-1.5 font-display text-xl font-medium text-ivory">
                Delete this workspace?
              </h2>
            </div>
            <p className="mb-6 text-sm text-sand-500">
              <span className="text-ivory">{workspace.name}</span> and all of
              its teams, memberships, and shared secrets will be permanently
              destroyed. This cannot be undone.
            </p>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setIsDeleteOpen(false)}
                className="lk-btn lk-btn--ghost"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                disabled={isDeleting}
                className="lk-btn lk-btn--danger"
              >
                {isDeleting ? "Deleting..." : "Delete Workspace"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}