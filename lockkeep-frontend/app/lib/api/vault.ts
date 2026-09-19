import { config } from "~/config";
import type { KDFParams, VaultItem, CreateVaultItemRequest, UpdateVaultItemRequest } from "~/types/index";
import { authFetch } from "./core";

export async function verifyVaultPassword(
    verificationHash: string,
): Promise<{ success: boolean; credentials: VaultItem[] }> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/vault/verify`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ verification_hash: verificationHash }),
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to verify vault password");
    }

    return res.json();
}

export async function createVaultItem(
    item: CreateVaultItemRequest,
): Promise<VaultItem> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/vault/item`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(item),
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to add item to vault");
    }

    return res.json();
}

export async function fetchVaultItems(): Promise<VaultItem[]> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/vault/items`, {
        method: "GET",
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to fetch vault items");
    }

    return res.json();
}

export async function updateVaultItem(
    itemId: string,
    updates: UpdateVaultItemRequest,
    updatePolicy?: boolean,
): Promise<VaultItem> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/vault/item/${itemId}?updatePolicy=${updatePolicy === true}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updates),
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to update vault item");
    }

    return res.json();
}

export async function deleteVaultItem(itemId: string): Promise<void> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/vault/item/${itemId}`, {
        method: "DELETE",
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to delete vault item");
    }
}

export async function updateVaultPassword(
    newVerificationHash: string,
    newKdfParams: KDFParams,
): Promise<void> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/auth/vault/update`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            verification_hash: newVerificationHash,
            kdf_params: newKdfParams,
        }),
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to update vault password");
    }
}