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
    hasMasterPassword: boolean;
    authMethod?: string;
    vault?: VaultMetadata;
    userType: "admin" | "user";   // NEW
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