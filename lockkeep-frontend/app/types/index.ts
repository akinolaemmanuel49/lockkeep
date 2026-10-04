// ─── Domain Types ─────────────────────────────────────────────

import { setCurrentPolicy } from "~/lib/api/crypto.policy";
import type { KDFAlgorithm } from "~/lib/crypto";

export type VaultItemType =
    | "login"
    | "environment"
    | "ssh_key"
    | "secure_note"
    | "payment_card"
    | "api_key";

export interface KDFParams {
    algorithm: "argon2id" | "scrypt";
    salt: string;
    memory: number;
    iterations: number;
    parallelism: number;
}

export interface VaultMetadata {
    version: number;
    verificationHash: string;
    kdf: KDFParams;
}

export interface User {
    id: string;
    email: string;
    tenantId: string;
    systemRole: string;
    hasMasterPassword: boolean;
    authMethod?: string;
    vault?: VaultMetadata;
}

export interface Secret {
    ciphertext: string;
    iv: string;
    tag: string;
    version?: number;
}

export interface VaultItem {
    id: string;
    userId: string;
    tenantId: string;
    type: VaultItemType;
    name: string;
    metadata?: Record<string, unknown>;
    secret: Secret;
    createdAt: string;
    updatedAt: string;
}

// ─── Legacy Credential (kept for gradual migration if needed) ───
// Re-export VaultItem as Credential for components not yet migrated
export type Credential = VaultItem;

// ─── Auth DTOs ──────────────────────────────────────────────────

export interface LocalRegisterRequest {
    username: string;
    email: string;
    password: string;
}

export interface LocalLoginRequest {
    email: string;
    password: string;
}

export interface SetVerificationHashRequest {
    verification_hash: string;
    kdf_params: KDFParams;
}

// ─── Vault DTOs ─────────────────────────────────────────────────

export interface CreateVaultItemRequest {
    type: VaultItemType;
    name: string;
    metadata?: Record<string, unknown>;
    secret: Secret;
}

export interface UpdateVaultItemRequest {
    type: VaultItemType;
    name: string;
    metadata?: Record<string, unknown>;
    secret: Secret;
}

/** User-supplied payload for creating or updating a vault item (secret in plaintext). */
export interface VaultItemInput {
    type: VaultItemType;
    name: string;
    metadata: Record<string, unknown>;
    secret: string;
}

export interface MigrateVaultRequest {
    expected_version: number;
    verification_hash: string;
    kdf_params: KDFParams;
    vault_updates: Array<{
        id: string;
        type: VaultItemType;
        name: string;
        secret: Secret;
    }>;
}

export interface CryptoPolicy {
    id: string,
    version: number,
    kdfParams: {
        algorithm: "argon2id" | "scrypt";
        memory: number;
        iterations: number;
        parallelism: number;
    },
    updatedAt: string,
}

export interface SetCurrentPolicyRequest {
    kdfParams: {
        algorithm: KDFAlgorithm;
        memory: number;
        iterations: number;
        parallelism: number;
    }
}

// ─── Workspace (organization) DTOs ───────────────────────────────

export interface WorkspaceSettings {
    require2FA: boolean;
    defaultTeamRole: string;
}

export interface Workspace {
    id: string;
    name: string;
    slug: string;
    ownerId: string;
    settings: WorkspaceSettings;
    createdAt: string;
    updatedAt: string;
}

export interface WorkspaceTeam {
    id: string;
    organizationId: string;
    name: string;
    slug: string;
    createdAt: string;
    updatedAt: string;
}

// ─── Workspace shared secrets DTOs ──────────────────────────────

export type SharedSecretKind = "env" | "note" | "file" | "json";

export interface WorkspaceApplication {
    id: string;
    organizationId: string;
    name: string;
    slug: string;
    description?: string;
    createdBy: string;
    createdAt: string;
    updatedAt: string;
}

export interface WorkspaceEnvironment {
    id: string;
    organizationId: string;
    applicationId: string;
    name: string;
    slug: string;
    isProtected: boolean;
    createdBy: string;
    createdAt: string;
    updatedAt: string;
}

/**
 * A workspace secret as seen by the authenticated human API path: the server
 * unwraps the DEK and returns the plaintext value. `value` is omitted on list
 * responses only if the server suppresses it (currently always present).
 */
export interface WorkspaceEnvSecret {
    key: string;
    kind: SharedSecretKind;
    version: number;
    value?: string;
    createdAt: string;
    updatedAt: string;
}

export interface CreateApplicationInput {
    name: string;
    slug: string;
    description?: string;
}

export interface CreateEnvironmentInput {
    name: string;
    slug: string;
    isProtected: boolean;
}

export interface CreateSecretInput {
    key: string;
    kind: SharedSecretKind;
    value: string;
}

export interface UpdateSecretInput {
    kind: SharedSecretKind;
    value: string;
}

// ─── Machine access (service accounts + API keys) DTOs ──────────

export type ApiKeyScope = "secrets:read" | "secrets:write";

export interface ServiceAccount {
    id: string;
    organizationId: string;
    name: string;
    description?: string;
    createdBy: string;
    disabledAt?: string;
    createdAt: string;
    updatedAt: string;
}

export interface ApiKeyRecord {
    id: string;
    serviceAccountId: string;
    environmentId: string;
    keyPrefix: string;
    scopes: ApiKeyScope[];
    createdBy: string;
    expiresAt?: string;
    lastUsedAt?: string;
    revokedAt?: string;
    createdAt: string;
    updatedAt: string;
}

/** The one-time mint response: raw key is only ever shown/returned here. */
export interface MintedApiKey {
    apiKey: string;
    key: ApiKeyRecord;
}

export interface MachineEnvelope {
    environmentId: string;
    kdf: {
        algorithm: string;
        salt: string;
        info: string;
        length: number;
    };
    wrappedDek: {
        keyId: string;
        algorithm: string;
        ciphertext: string;
        nonce: string;
        aad: string;
    };
    secrets: Array<{
        key: string;
        kind: SharedSecretKind;
        version: number;
        ciphertext: string;
        nonce: string;
    }>;
}

export interface DecryptedMachineSecret {
    key: string;
    kind: SharedSecretKind;
    version: number;
    value: string;
}