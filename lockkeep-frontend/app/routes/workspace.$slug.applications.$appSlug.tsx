import { useEffect, useState } from "react";
import { Link, useParams } from "react-router";
import { requireAuth } from "~/lib/auth-guard";
import type { Route } from "./+types/workspace.$slug.applications.$appSlug";
import { fetchApplication } from "~/lib/api/sharedSecrets";
import type { WorkspaceApplication } from "~/types";
import Nameplate from "~/components/Nameplate";
import ApplicationSecrets from "~/components/ApplicationSecrets";

export const clientLoader = () => {
  return requireAuth();
};

export function meta({}: Route.MetaArgs) {
  return [{ title: "Application — LockKeep" }];
}

export default function ApplicationDetail() {
  const { slug, appSlug } = useParams();

  const [app, setApp] = useState<WorkspaceApplication | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!slug || !appSlug) return;
    let cancelled = false;

    (async () => {
      setIsLoading(true);
      setLoadError(null);
      try {
        const loaded = await fetchApplication(slug, appSlug);
        if (cancelled) return;
        setApp(loaded);
      } catch (err) {
        if (!cancelled) {
          setLoadError(
            err instanceof Error ? err.message : "Failed to load application",
          );
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [slug, appSlug]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center gap-3 py-24 text-sand-500">
        <div className="h-5 w-5 animate-spin rounded-full border-2 border-brass-600 border-t-brass-400" />
        <span>Opening the application...</span>
      </div>
    );
  }

  if (loadError || !app) {
    return (
      <div className="mx-auto max-w-lg py-16">
        <div className="plate p-8 text-center">
          <p className="text-sm text-vermillion-300">
            {loadError ?? "Application not found"}
          </p>
          {slug && (
            <Link to={`/workspaces/${slug}`} className="lk-btn lk-btn--ghost mt-6">
              ← Back to Workspace
            </Link>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl">
      {slug && (
        <Link
          to={`/workspaces/${slug}`}
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
          Back to Workspace
        </Link>
      )}

      <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <Nameplate>Application</Nameplate>
          <h1 className="mt-2 font-display text-3xl font-semibold tracking-tight text-ivory">
            {app.name}
          </h1>
          <p className="mt-1 font-mono text-sm text-brass-400/80">
            {app.slug}
            {app.description ? ` — ${app.description}` : ""}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.2em] text-sand-500">
            <span className="h-2 w-2 rounded-full bg-patina-400" />
            Environments & secrets
          </span>
        </div>
      </div>

      {slug && <ApplicationSecrets orgSlug={slug} app={app} />}
    </div>
  );
}