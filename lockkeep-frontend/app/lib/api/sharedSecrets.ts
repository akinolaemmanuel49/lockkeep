import { config } from "~/config";
import type {
    CreateApplicationInput,
    CreateEnvironmentInput,
    CreateSecretInput,
    UpdateSecretInput,
    WorkspaceApplication,
    WorkspaceEnvironment,
    WorkspaceEnvSecret,
} from "~/types";
import { authFetch } from "./core";

const orgBase = (slug: string) =>
    `${config.LOCKKEEP_API_URI}/organizations/${encodeURIComponent(slug)}`;

const appBase = (orgSlug: string, appSlug: string) =>
    `${orgBase(orgSlug)}/applications/${encodeURIComponent(appSlug)}`;

const envBase = (orgSlug: string, appSlug: string, envSlug: string) =>
    `${appBase(orgSlug, appSlug)}/environments/${encodeURIComponent(envSlug)}`;

async function asError(res: Response, fallback: string): Promise<Error> {
    try {
        const body = await res.json();
        return new Error(body.error || fallback);
    } catch {
        return new Error(fallback);
    }
}

// ─── Applications ────────────────────────────────────────────────

export async function fetchApplications(
    slug: string,
): Promise<WorkspaceApplication[]> {
    const res = await authFetch(`${orgBase(slug)}/applications`, {
        method: "GET",
    });
    if (!res.ok) throw await asError(res, "Failed to fetch applications");
    const data = await res.json();
    return data.applications ?? [];
}

export async function createApplication(
    slug: string,
    input: CreateApplicationInput,
): Promise<WorkspaceApplication> {
    const res = await authFetch(`${orgBase(slug)}/applications`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(input),
    });
    if (!res.ok) throw await asError(res, "Failed to create application");
    const data = await res.json();
    return data.application;
}

export async function fetchApplication(
    orgSlug: string,
    appSlug: string,
): Promise<WorkspaceApplication> {
    const res = await authFetch(`${appBase(orgSlug, appSlug)}`, {
        method: "GET",
    });
    if (!res.ok) throw await asError(res, "Failed to fetch application");
    const data = await res.json();
    return data.application;
}

export async function deleteApplication(
    orgSlug: string,
    appSlug: string,
): Promise<void> {
    const res = await authFetch(`${appBase(orgSlug, appSlug)}`, {
        method: "DELETE",
    });
    if (!res.ok) throw await asError(res, "Failed to delete application");
}

// ─── Environments ────────────────────────────────────────────────

export async function fetchEnvironments(
    orgSlug: string,
    appSlug: string,
): Promise<WorkspaceEnvironment[]> {
    const res = await authFetch(`${appBase(orgSlug, appSlug)}/environments`, {
        method: "GET",
    });
    if (!res.ok) throw await asError(res, "Failed to fetch environments");
    const data = await res.json();
    return data.environments ?? [];
}

export async function createEnvironment(
    orgSlug: string,
    appSlug: string,
    input: CreateEnvironmentInput,
): Promise<WorkspaceEnvironment> {
    const res = await authFetch(`${appBase(orgSlug, appSlug)}/environments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(input),
    });
    if (!res.ok) throw await asError(res, "Failed to create environment");
    const data = await res.json();
    return data.environment;
}

export async function deleteEnvironment(
    orgSlug: string,
    appSlug: string,
    envSlug: string,
): Promise<void> {
    const res = await authFetch(`${envBase(orgSlug, appSlug, envSlug)}`, {
        method: "DELETE",
    });
    if (!res.ok) throw await asError(res, "Failed to delete environment");
}

// ─── Secrets ─────────────────────────────────────────────────────

export async function fetchSecrets(
    orgSlug: string,
    appSlug: string,
    envSlug: string,
): Promise<WorkspaceEnvSecret[]> {
    const res = await authFetch(
        `${envBase(orgSlug, appSlug, envSlug)}/secrets`,
        { method: "GET" },
    );
    if (!res.ok) throw await asError(res, "Failed to fetch secrets");
    const data = await res.json();
    return data.secrets ?? [];
}

export async function createSecret(
    orgSlug: string,
    appSlug: string,
    envSlug: string,
    input: CreateSecretInput,
): Promise<WorkspaceEnvSecret> {
    const res = await authFetch(
        `${envBase(orgSlug, appSlug, envSlug)}/secrets`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(input),
        },
    );
    if (!res.ok) throw await asError(res, "Failed to save secret");
    const data = await res.json();
    return data.secret;
}

export async function updateSecret(
    orgSlug: string,
    appSlug: string,
    envSlug: string,
    key: string,
    input: UpdateSecretInput,
): Promise<WorkspaceEnvSecret> {
    const res = await authFetch(
        `${envBase(orgSlug, appSlug, envSlug)}/secrets/${encodeURIComponent(key)}`,
        {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(input),
        },
    );
    if (!res.ok) throw await asError(res, "Failed to save secret");
    const data = await res.json();
    return data.secret;
}

export async function deleteSecret(
    orgSlug: string,
    appSlug: string,
    envSlug: string,
    key: string,
): Promise<void> {
    const res = await authFetch(
        `${envBase(orgSlug, appSlug, envSlug)}/secrets/${encodeURIComponent(key)}`,
        { method: "DELETE" },
    );
    if (!res.ok) throw await asError(res, "Failed to delete secret");
}