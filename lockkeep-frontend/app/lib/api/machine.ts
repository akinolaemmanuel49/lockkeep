import { config } from "~/config";
import type {
    DecryptedMachineSecret,
    MachineEnvelope,
} from "~/types";

// Browser-side mirror of the @lockkeep/sdk unwrap pipeline. The web console
// proves the flow the SDK automates: derive the wrapping key from the raw API
// key (HKDF-SHA256), unwrap the environment data key (AES-GCM), then decrypt
// every sealed entry. `kdf.salt` and `wrappedDek.aad` are the environment-id
// hex STRING used as UTF-8 bytes — never hex-decoded.

const utf8 = (value: string): Uint8Array<ArrayBuffer> =>
    new TextEncoder().encode(value);

function base64ToBytes(b64: string): Uint8Array<ArrayBuffer> {
    const binary = atob(b64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
        bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
}

/** Fetches the raw ciphertext envelope for the environment bound to the key. */
export async function fetchMachineEnvelope(
    apiKey: string,
): Promise<MachineEnvelope> {
    const res = await fetch(`${config.LOCKKEEP_API_URI}/machine/secrets`, {
        method: "GET",
        headers: { Authorization: `Bearer ${apiKey}` },
    });
    if (!res.ok) {
        let message = `Machine export failed: HTTP ${res.status}`;
        try {
            const body = await res.json();
            if (body.error) message = body.error;
        } catch {
            // keep fallback
        }
        throw new Error(message);
    }
    return (await res.json()) as MachineEnvelope;
}

async function unwrapEnvDek(
    envelope: MachineEnvelope,
    apiKey: string,
): Promise<CryptoKey> {
    const material = await crypto.subtle.importKey(
        "raw",
        utf8(apiKey),
        "HKDF",
        false,
        ["deriveBits"],
    );
    const derived = new Uint8Array(
        await crypto.subtle.deriveBits(
            {
                name: "HKDF",
                hash: "SHA-256",
                salt: utf8(envelope.kdf.salt),
                info: utf8(envelope.kdf.info),
            },
            material,
            envelope.kdf.length * 8,
        ),
    );
    const kek = await crypto.subtle.importKey(
        "raw",
        derived,
        { name: "AES-GCM", length: 256 },
        false,
        ["decrypt"],
    );

    let unwrapped: ArrayBuffer;
    try {
        unwrapped = await crypto.subtle.decrypt(
            {
                name: "AES-GCM",
                iv: base64ToBytes(envelope.wrappedDek.nonce),
                additionalData: utf8(envelope.wrappedDek.aad),
            },
            kek,
            base64ToBytes(envelope.wrappedDek.ciphertext),
        );
    } catch {
        throw new Error(
            "Failed to unwrap the environment key — wrong API key or tampered data",
        );
    }

    const dek = new Uint8Array(unwrapped);
    try {
        return await crypto.subtle.importKey(
            "raw",
            dek,
            { name: "AES-GCM", length: 256 },
            false,
            ["decrypt"],
        );
    } finally {
        derived.fill(0);
        dek.fill(0);
    }
}

async function decryptEntry(
    dek: CryptoKey,
    entry: MachineEnvelope["secrets"][number],
    environmentId: string,
): Promise<string> {
    const aad = utf8(`${environmentId}|${entry.key}|${entry.version}`);
    const plaintext = await crypto.subtle.decrypt(
        {
            name: "AES-GCM",
            iv: base64ToBytes(entry.nonce),
            additionalData: aad,
        },
        dek,
        base64ToBytes(entry.ciphertext),
    );
    return new TextDecoder().decode(plaintext);
}

/**
 * Fetches the envelope with the raw key and decrypts every secret in the
 * browser. Used by the Machine Access panel to prove the environment key works.
 */
export async function fetchAndDecryptEnvelope(
    apiKey: string,
): Promise<{ environmentId: string; secrets: DecryptedMachineSecret[] }> {
    const envelope = await fetchMachineEnvelope(apiKey);
    const dek = await unwrapEnvDek(envelope, apiKey);
    const secrets: DecryptedMachineSecret[] = [];
    for (const entry of envelope.secrets) {
        secrets.push({
            key: entry.key,
            kind: entry.kind,
            version: entry.version,
            value: await decryptEntry(dek, entry, envelope.environmentId),
        });
    }
    return { environmentId: envelope.environmentId, secrets };
}