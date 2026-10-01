import { describe, expect, it, vi } from "vitest";
import { fetchAndDecryptEnvelope } from "./machine";
import type { MachineEnvelope } from "~/types";

const utf8 = (value: string): Uint8Array<ArrayBuffer> =>
    new TextEncoder().encode(value);
const b64 = (bytes: Uint8Array) => Buffer.from(bytes).toString("base64");

async function aesGcmEncrypt(
    key: Uint8Array<ArrayBuffer>,
    plaintext: Uint8Array<ArrayBuffer>,
    aad: Uint8Array<ArrayBuffer>,
): Promise<{ ciphertext: string; nonce: string }> {
    const imported = await crypto.subtle.importKey(
        "raw",
        key,
        { name: "AES-GCM", length: 256 },
        false,
        ["encrypt"],
    );
    const nonce = crypto.getRandomValues(new Uint8Array(12));
    const sealed = new Uint8Array(
        await crypto.subtle.encrypt(
            { name: "AES-GCM", iv: nonce, additionalData: aad },
            imported,
            plaintext,
        ),
    );
    return { ciphertext: b64(sealed), nonce: b64(nonce) };
}

async function buildEnvelope(): Promise<MachineEnvelope> {
    const envId = "65b3f7edd9bfca00daa6e3b31";
    const apiKey = "lk_live_" + "cd".repeat(24);
    const dek = crypto.getRandomValues(new Uint8Array(32));

    const material = await crypto.subtle.importKey(
        "raw",
        utf8(apiKey),
        "HKDF",
        false,
        ["deriveBits"],
    );
    const kek = new Uint8Array(
        await crypto.subtle.deriveBits(
            {
                name: "HKDF",
                hash: "SHA-256",
                salt: utf8(envId),
                info: utf8("lockkeep:client-dek-wrap"),
            },
            material,
            32 * 8,
        ),
    );

    const wrappedDek = await aesGcmEncrypt(kek, dek, utf8(envId));

    const value = "postgres://u:p@db:5432/app";
    const entry = await aesGcmEncrypt(
        dek,
        utf8(value),
        utf8(`${envId}|DATABASE_URL|1`),
    );

    return {
        environmentId: envId,
        kdf: {
            algorithm: "hkdf-sha256",
            salt: envId,
            info: "lockkeep:client-dek-wrap",
            length: 32,
        },
        wrappedDek: {
            keyId: "65b3f7edd9bfca00daa6e3b32",
            algorithm: "aes-256-gcm-v1",
            ...wrappedDek,
            aad: envId,
        },
        secrets: [
            {
                key: "DATABASE_URL",
                kind: "env",
                version: 1,
                ...entry,
            },
        ],
    };
}

describe("fetchAndDecryptEnvelope", () => {
    it("fetches the envelope and decrypts the secret in the browser mirror", async () => {
        const envelope = await buildEnvelope();
        vi.stubGlobal(
            "fetch",
            vi.fn(async () =>
                new Response(JSON.stringify(envelope), {
                    status: 200,
                    headers: { "Content-Type": "application/json" },
                }),
            ),
        );

        const apiKey = "lk_live_" + "cd".repeat(24);
        const result = await fetchAndDecryptEnvelope(apiKey);

        expect(result.environmentId).toBe(envelope.environmentId);
        expect(result.secrets).toEqual([
            {
                key: "DATABASE_URL",
                kind: "env",
                version: 1,
                value: "postgres://u:p@db:5432/app",
            },
        ]);
        vi.unstubAllGlobals();
    });
});