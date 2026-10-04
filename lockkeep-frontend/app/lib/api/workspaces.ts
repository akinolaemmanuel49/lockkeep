import { config } from "~/config";
import type { Workspace, WorkspaceTeam } from "~/types";
import { authFetch } from "./core";

export async function fetchWorkspaces(): Promise<Workspace[]> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/organizations`, {
        method: "GET",
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to fetch workspaces");
    }

    const data = await res.json();
    return data.organizations ?? [];
}

export async function createWorkspace(input: {
    name: string;
    slug: string;
}): Promise<Workspace> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/organizations`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(input),
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to create workspace");
    }

    const data = await res.json();
    return data.organization;
}

export async function fetchWorkspace(slug: string): Promise<Workspace> {
    const res = await authFetch(
        `${config.LOCKKEEP_API_URI}/organizations/${encodeURIComponent(slug)}`,
        { method: "GET" },
    );

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to fetch workspace");
    }

    const data = await res.json();
    return data.organization;
}

export async function fetchWorkspaceTeams(slug: string): Promise<WorkspaceTeam[]> {
    const res = await authFetch(
        `${config.LOCKKEEP_API_URI}/organizations/${encodeURIComponent(slug)}/teams`,
        { method: "GET" },
    );

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to fetch teams");
    }

    const data = await res.json();
    return data.teams ?? [];
}

export async function createWorkspaceTeam(
    slug: string,
    input: { name: string; slug: string },
): Promise<WorkspaceTeam> {
    const res = await authFetch(
        `${config.LOCKKEEP_API_URI}/organizations/${encodeURIComponent(slug)}/teams`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(input),
        },
    );

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to create team");
    }

    const data = await res.json();
    return data.team;
}

export async function updateWorkspace(
    slug: string,
    input: { name?: string; slug?: string },
): Promise<Workspace> {
    const res = await authFetch(
        `${config.LOCKKEEP_API_URI}/organizations/${encodeURIComponent(slug)}`,
        {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(input),
        },
    );

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to update workspace");
    }

    const data = await res.json();
    return data.organization;
}

export async function deleteWorkspace(slug: string): Promise<void> {
    const res = await authFetch(
        `${config.LOCKKEEP_API_URI}/organizations/${encodeURIComponent(slug)}`,
        { method: "DELETE" },
    );

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to delete workspace");
    }
}