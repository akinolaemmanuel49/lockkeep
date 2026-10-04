import { config } from "~/config";
import type {
    ApiKeyRecord,
    ApiKeyScope,
    MintedApiKey,
    ServiceAccount,
} from "~/types";
import { authFetch } from "./core";

const orgBase = (slug: string) =>
    `${config.LOCKKEEP_API_URI}/organizations/${encodeURIComponent(slug)}`;

const accountsBase = (orgSlug: string) =>
    `${orgBase(orgSlug)}/service-accounts`;

const accountBase = (orgSlug: string, accountId: string) =>
    `${accountsBase(orgSlug)}/${encodeURIComponent(accountId)}`;

async function asError(res: Response, fallback: string): Promise<Error> {
    try {
        const body = await res.json();
        return new Error(body.error || fallback);
    } catch {
        return new Error(fallback);
    }
}

export async function fetchServiceAccounts(
    orgSlug: string,
): Promise<ServiceAccount[]> {
    const res = await authFetch(accountsBase(orgSlug), { method: "GET" });
    if (!res.ok) throw await asError(res, "Failed to fetch service accounts");
    const data = await res.json();
    return data.serviceAccounts ?? [];
}

export async function createServiceAccount(
    orgSlug: string,
    input: { name: string; description?: string },
): Promise<ServiceAccount> {
    const res = await authFetch(accountsBase(orgSlug), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(input),
    });
    if (!res.ok) throw await asError(res, "Failed to create service account");
    const data = await res.json();
    return data.serviceAccount;
}

export async function deleteServiceAccount(
    orgSlug: string,
    accountId: string,
): Promise<void> {
    const res = await authFetch(accountBase(orgSlug, accountId), {
        method: "DELETE",
    });
    if (!res.ok)
        throw await asError(res, "Failed to delete service account");
}

export async function fetchApiKeys(
    orgSlug: string,
    accountId: string,
): Promise<ApiKeyRecord[]> {
    const res = await authFetch(
        `${accountBase(orgSlug, accountId)}/api-keys`,
        { method: "GET" },
    );
    if (!res.ok) throw await asError(res, "Failed to fetch API keys");
    const data = await res.json();
    return data.apiKeys ?? [];
}

export async function mintApiKey(
    orgSlug: string,
    accountId: string,
    input: {
        applicationSlug: string;
        environmentSlug: string;
        scopes: ApiKeyScope[];
        expiresAt?: string;
    },
): Promise<MintedApiKey> {
    const res = await authFetch(
        `${accountBase(orgSlug, accountId)}/api-keys`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(input),
        },
    );
    if (!res.ok) throw await asError(res, "Failed to mint API key");
    const data = await res.json();
    return { apiKey: data.apiKey, key: data.key };
}

export async function revokeApiKey(
    orgSlug: string,
    accountId: string,
    keyId: string,
): Promise<void> {
    const res = await authFetch(
        `${accountBase(orgSlug, accountId)}/api-keys/${encodeURIComponent(keyId)}`,
        { method: "DELETE" },
    );
    if (!res.ok) throw await asError(res, "Failed to revoke API key");
}