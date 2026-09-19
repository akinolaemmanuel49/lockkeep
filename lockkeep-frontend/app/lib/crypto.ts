import { argon2idAsync } from "@noble/hashes/argon2.js";
import { scryptAsync } from "@noble/hashes/scrypt.js";
import { sha256 } from "@noble/hashes/sha2.js";
import { bytesToHex, hexToBytes, randomBytes } from "@noble/hashes/utils.js";
import type { VaultMetadata } from "../types/index";

// ─── Types ─────────────────────────────────────────────────────────────────

export type KDFAlgorithm = "argon2id" | "scrypt";

export interface KDFParams {
    algorithm: KDFAlgorithm;
    salt: string;        // hex
    memory: number;      // KB
    iterations: number;
    parallelism: number;
}

export interface EncryptionResult {
    ciphertext: string;
    iv: string;
    tag: string;
    version?: number; // The version number comes from the server-stored VaultItem.secret.version
    // for decryption, or from CURRENT_VERSION when creating new items.
}

// ─── KDF Resolution ────────────────────────────────────────────────────────

/**
 * Derive encryption key and verification hash from vault password.
 * Uses the KDF params stored server-side (algorithm, memory, iterations, etc).
 */
export async function deriveKeys(
    password: string,
    params: KDFParams,
): Promise<{
    encryptionKey: CryptoKey;
    verificationHash: string;
}> {
    const passwordBytes = new TextEncoder().encode(password);

    const salt = hexToBytes(params.salt);


    let derived: Uint8Array;

    switch (params.algorithm) {
        case "argon2id":
            derived = await argon2idAsync(passwordBytes, salt, {
                t: params.iterations,
                m: params.memory,
                p: params.parallelism,
                dkLen: 64,
            });
            break;

        case "scrypt":
            derived = await scryptAsync(passwordBytes, salt, {
                N: params.memory * 1024,
                r: 8,
                p: params.parallelism,
                dkLen: 64,
            });
            break;

        default:
            throw new Error("Unsupported algorithm");
    }

    const enc = derived.slice(0, 32);
    const verify = derived.slice(32);

    const encryptionKey = await crypto.subtle.importKey(
        "raw",
        enc,
        { name: "AES-GCM", length: 256 },
        false,
        ["encrypt", "decrypt"],
    );

    const verificationHash = bytesToHex(sha256(verify));

    derived.fill(0);
    enc.fill(0);
    verify.fill(0);

    return {
        encryptionKey,
        verificationHash,
    };
}

// ─── AES-GCM Encryption ────────────────────────────────────────────────────

export async function encryptSecret(
    plaintext: string,
    key: CryptoKey,
    version?: number,
): Promise<EncryptionResult> {
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const data = new TextEncoder().encode(plaintext);

    const ciphertext = await crypto.subtle.encrypt(
        { name: "AES-GCM", iv },
        key,
        data,
    );

    const full = new Uint8Array(ciphertext);
    const encrypted = full.slice(0, full.length - 16);
    const tag = full.slice(full.length - 16);

    return {
        ciphertext: bytesToHex(encrypted),
        iv: bytesToHex(iv),
        tag: bytesToHex(tag),
    };
}

export async function decryptSecret(
    result: EncryptionResult,
    key: CryptoKey,
): Promise<string> {
    const encryptedBytes = hexToBytes(result.ciphertext);
    const ivBytes = hexToBytes(result.iv);
    const tagBytes = hexToBytes(result.tag);

    const full = new Uint8Array(encryptedBytes.length + tagBytes.length);
    full.set(encryptedBytes, 0);
    full.set(tagBytes, encryptedBytes.length);

    const decrypted = await crypto.subtle.decrypt(
        { name: "AES-GCM", iv: ivBytes },
        key,
        full,
    );

    return new TextDecoder().decode(decrypted);
}

// ─── Helpers ──────────────────────────────────────────────────────────────

export function generateSalt(): Uint8Array {
    return randomBytes(32);
}

/**
 * Builds the KDF parameters for a vault using the
 * server-provided vault metadata and a newly generated salt.
 */
export function buildKDFParams(
    kdfParams: KDFParams,
    salt: Uint8Array,
): KDFParams {
    return {
        algorithm: kdfParams.algorithm,
        memory: kdfParams.memory,
        iterations: kdfParams.iterations,
        parallelism: kdfParams.parallelism,
        salt: bytesToHex(salt),
    };
}