import { describe, it, expect } from "vitest";
import {
    deriveKeys,
    encryptSecret,
    decryptSecret,
    generateSalt,
    buildKDFParams,
    type KDFParams,
} from "./crypto";

const scryptParams: KDFParams = {
    algorithm: "scrypt",
    salt: "0e5a171496fe3bb32a53c49663c8539d14a703a73568cde3903a3e89d49047d8",
    memory: 16,
    iterations: 2,
    parallelism: 1,
};

const argon2Params: KDFParams = {
    algorithm: "argon2id",
    salt: "0e5a171496fe3bb32a53c49663c8539d14a703a73568cde3903a3e89d49047d8",
    memory: 4096,
    iterations: 2,
    parallelism: 1,
};

describe("deriveKeys", () => {
    it("produces a 64-char hex verification hash and a usable AES-GCM key for scrypt", async () => {
        const { encryptionKey, verificationHash } = await deriveKeys("correct horse", scryptParams);

        expect(verificationHash).toMatch(/^[0-9a-f]{64}$/);
        expect(encryptionKey.algorithm.name).toBe("AES-GCM");
    });

    it("supports argon2id", async () => {
        const { verificationHash } = await deriveKeys("correct horse", argon2Params);
        expect(verificationHash).toMatch(/^[0-9a-f]{64}$/);
    });

    it("is deterministic for the same password and params", async () => {
        const a = await deriveKeys("same-password", scryptParams);
        const b = await deriveKeys("same-password", scryptParams);
        expect(a.verificationHash).toBe(b.verificationHash);
    });

    it("changes when the password changes", async () => {
        const a = await deriveKeys("password-a", scryptParams);
        const b = await deriveKeys("password-b", scryptParams);
        expect(a.verificationHash).not.toBe(b.verificationHash);
    });

    it("changes when the salt changes", async () => {
        const params = { ...scryptParams, salt: "0000000000000000000000000000000000000000000000000000000000000000" };
        const a = await deriveKeys("same-password", scryptParams);
        const b = await deriveKeys("same-password", params);
        expect(a.verificationHash).not.toBe(b.verificationHash);
    });

    it("rejects an unsupported algorithm", async () => {
        await expect(deriveKeys("pw", { ...scryptParams, algorithm: "pbkdf2" as never })).rejects.toThrow(
            "Unsupported algorithm",
        );
    });
});

describe("encryptSecret / decryptSecret round trip", () => {
    for (const params of [scryptParams, argon2Params]) {
        it(`round-trips plaintext (${params.algorithm})`, async () => {
            const { encryptionKey } = await deriveKeys("test-password", params);
            const plaintext = "hunter2-secret-value";

            const encrypted = await encryptSecret(plaintext, encryptionKey);
            expect(encrypted.ciphertext).toBeTruthy();
            expect(encrypted.iv).toBeTruthy();
            expect(encrypted.tag).toBeTruthy();

            const decrypted = await decryptSecret(encrypted, encryptionKey);
            expect(decrypted).toBe(plaintext);
        });
    }

    it("produces a fresh iv and ciphertext on every encryption", async () => {
        const { encryptionKey } = await deriveKeys("test-password", scryptParams);

        const first = await encryptSecret("same plaintext", encryptionKey);
        const second = await encryptSecret("same plaintext", encryptionKey);

        expect(first.iv).not.toBe(second.iv);
        expect(first.ciphertext).not.toBe(second.ciphertext);
    });

    it("fails to decrypt with the wrong key", async () => {
        const keyA = (await deriveKeys("password-a", scryptParams)).encryptionKey;
        const keyB = (await deriveKeys("password-b", scryptParams)).encryptionKey;

        const encrypted = await encryptSecret("secret", keyA);
        await expect(decryptSecret(encrypted, keyB)).rejects.toThrow();
    });
});

describe("generateSalt", () => {
    it("returns 32 random bytes", () => {
        const salt = generateSalt();
        expect(salt).toHaveLength(32);
    });

    it("generates distinct salts", () => {
        expect(generateSalt()).not.toEqual(generateSalt());
    });
});

describe("buildKDFParams", () => {
    it("keeps algorithm/cost params and swaps in the new salt", () => {
        const params = buildKDFParams(scryptParams, generateSalt());
        expect(params.algorithm).toBe("scrypt");
        expect(params.memory).toBe(scryptParams.memory);
        expect(params.iterations).toBe(scryptParams.iterations);
        expect(params.parallelism).toBe(scryptParams.parallelism);
        expect(params.salt).not.toBe(scryptParams.salt);
        expect(params.salt).toMatch(/^[0-9a-f]{64}$/);
    });
});