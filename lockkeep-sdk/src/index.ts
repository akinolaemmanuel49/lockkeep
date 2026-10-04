// ─── LockKeep machine-export client ────────────────────────────────────────
//
// The machine export endpoint (`GET /api/v2/machine/secrets`) only ever
// returns ciphertext: the per-environment data envelope is wrapped so that the
// server itself cannot read the values during transport. A client with the
// raw API key derives the wrapping key and unwraps the environment data key,
// then decrypts each secret.
//
// Conventions (mirrored from the Go backend crypto package):
//   - HKDF-SHA256(ikm = raw API key bytes, salt = UTF-8 encoding of the
//     environment-id hex STRING, info = "lockkeep:client-dek-wrap", length 32)
//   - wrappedDek: AES-256-GCM(dek, aad = UTF-8 env-id hex string)
//   - each secret: AES-256-GCM(plaintext, aad = `${envIdHex}|${key}|${version}`)
//
// `kdf.salt` and `wrappedDek.aad` are text fields holding the 24-char
// environment-id hex string. They are used as UTF-8 bytes — never hex-decoded.

export const HKDF_INFO = "lockkeep:client-dek-wrap";
export const WRAP_ALGORITHM = "aes-256-gcm-v1";
export const DEFAULT_BASE_URL = "http://localhost:8000/api/v2";

export interface KdfInfo {
  algorithm: string;
  salt: string;
  info: string;
  length: number;
}

export interface WrappedDek {
  keyId: string;
  algorithm: string;
  ciphertext: string;
  nonce: string;
  aad: string;
}

export interface MachineSecretEntry {
  key: string;
  kind: string;
  version: number;
  ciphertext: string;
  nonce: string;
}

export interface MachineEnvelope {
  environmentId: string;
  kdf: KdfInfo;
  wrappedDek: WrappedDek;
  secrets: MachineSecretEntry[];
}

export interface DecryptedSecret {
  key: string;
  kind: string;
  version: number;
  value: string;
}

export interface SecretsSet {
  environmentId: string;
  secrets: DecryptedSecret[];
}

export type LockKeepErrorCode =
  | "http"
  | "unsupported-environment"
  | "unsupported-kdf"
  | "invalid-key"
  | "decryption-failed";

export class LockKeepError extends Error {
  readonly code: LockKeepErrorCode;
  readonly status?: number;

  constructor(
    message: string,
    options: { code: LockKeepErrorCode; status?: number } = {
      code: "decryption-failed",
    },
  ) {
    super(message);
    this.name = "LockKeepError";
    this.code = options.code;
    if (options.status !== undefined) this.status = options.status;
  }
}

// ─── Encoding helpers ──────────────────────────────────────────────────────

function utf8(value: string): Uint8Array<ArrayBuffer> {
  return new TextEncoder().encode(value);
}

function base64ToBytes(b64: string): Uint8Array<ArrayBuffer> {
  const binary = atob(b64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

function getSubtle(): SubtleCrypto {
  const subtle = globalThis.crypto?.subtle;
  if (!subtle) {
    throw new LockKeepError(
      "WebCrypto (crypto.subtle) is required; run on Node.js >= 18 or a modern browser",
      { code: "unsupported-environment" },
    );
  }
  return subtle;
}

// ─── Core crypto ───────────────────────────────────────────────────────────

/**
 * Derives the client wrapping key from the raw API key and the envelope's KDF
 * parameters, then unwraps the environment data key (DEK).
 */
export async function unwrapEnvelope(
  envelope: MachineEnvelope,
  apiKey: string,
): Promise<CryptoKey> {
  if (envelope.kdf.algorithm !== "hkdf-sha256") {
    throw new LockKeepError(
      `unsupported kdf algorithm: ${envelope.kdf.algorithm}`,
      { code: "unsupported-kdf" },
    );
  }
  if (envelope.kdf.info !== HKDF_INFO) {
    throw new LockKeepError(`unsupported kdf info: ${envelope.kdf.info}`, {
      code: "unsupported-kdf",
    });
  }
  if (envelope.kdf.length !== 32) {
    throw new LockKeepError(`unsupported kdf length: ${envelope.kdf.length}`, {
      code: "unsupported-kdf",
    });
  }
  if (envelope.wrappedDek.algorithm !== WRAP_ALGORITHM) {
    throw new LockKeepError(
      `unsupported wrap algorithm: ${envelope.wrappedDek.algorithm}`,
      { code: "unsupported-kdf" },
    );
  }

  const subtle = getSubtle();

  const material = await subtle.importKey(
    "raw",
    utf8(apiKey),
    "HKDF",
    false,
    ["deriveBits"],
  );

  const derived = await subtle.deriveBits(
    {
      name: "HKDF",
      hash: "SHA-256",
      salt: utf8(envelope.kdf.salt),
      info: utf8(envelope.kdf.info),
    },
    material,
    envelope.kdf.length * 8,
  );

  const kek = new Uint8Array(derived);
  const kekKey = await subtle.importKey(
    "raw",
    kek,
    { name: "AES-GCM", length: 256 },
    false,
    ["decrypt"],
  );

  let unwrapped: ArrayBuffer;
  try {
    unwrapped = await subtle.decrypt(
      {
        name: "AES-GCM",
        iv: base64ToBytes(envelope.wrappedDek.nonce),
        additionalData: utf8(envelope.wrappedDek.aad),
      },
      kekKey,
      base64ToBytes(envelope.wrappedDek.ciphertext),
    );
  } catch (err) {
    throw new LockKeepError(
      err instanceof Error && err.message
        ? `failed to unwrap environment key: ${err.message}`
        : "failed to unwrap environment key — wrong API key or tampered data",
      { code: "invalid-key" },
    );
  }

  const dek = new Uint8Array(unwrapped);
  try {
    return await subtle.importKey(
      "raw",
      dek,
      { name: "AES-GCM", length: 256 },
      false,
      ["decrypt"],
    );
  } catch (err) {
    throw new LockKeepError(
      err instanceof Error ? err.message : "invalid dek",
      { code: "invalid-key" },
    );
  } finally {
    dek.fill(0);
  }
}

/**
 * Decrypts a single export entry with the already-unwrapped DEK. The AAD is
 * always `${environmentId}|${key}|${version}`.
 */
export async function decryptSecret(
  dek: CryptoKey,
  entry: MachineSecretEntry,
  environmentId: string,
): Promise<string> {
  const subtle = getSubtle();
  const aad = utf8(`${environmentId}|${entry.key}|${entry.version}`);
  let plaintext: ArrayBuffer;
  try {
    plaintext = await subtle.decrypt(
      {
        name: "AES-GCM",
        iv: base64ToBytes(entry.nonce),
        additionalData: aad,
      },
      dek,
      base64ToBytes(entry.ciphertext),
    );
  } catch {
    throw new LockKeepError(
      `failed to decrypt "${entry.key}" — wrong API key, tampered data, or unknown version`,
      { code: "decryption-failed" },
    );
  }
  return new TextDecoder().decode(plaintext);
}

export async function decryptEnvelope(
  envelope: MachineEnvelope,
  apiKey: string,
): Promise<SecretsSet> {
  const dek = await unwrapEnvelope(envelope, apiKey);
  const secrets: DecryptedSecret[] = [];
  for (const entry of envelope.secrets) {
    secrets.push({
      key: entry.key,
      kind: entry.kind,
      version: entry.version,
      value: await decryptSecret(dek, entry, envelope.environmentId),
    });
  }
  return { environmentId: envelope.environmentId, secrets };
}

// ─── Fetching ──────────────────────────────────────────────────────────────

export interface FetchOptions {
  apiKey: string;
  /** LockKeep API base URL, e.g. "https://lockkeep.example.com/api/v2". */
  baseUrl?: string;
  signal?: AbortSignal;
}

/** Fetches the raw ciphertext envelope for the environment bound to the key. */
export async function fetchEnvelope(options: FetchOptions): Promise<MachineEnvelope> {
  const baseUrl = (options.baseUrl ?? DEFAULT_BASE_URL).replace(/\/+$/, "");
  let res: Response;
  try {
    res = await fetch(`${baseUrl}/machine/secrets`, {
      method: "GET",
      headers: { Authorization: `Bearer ${options.apiKey}` },
      signal: options.signal,
    });
  } catch (err) {
    throw new LockKeepError(
      err instanceof Error ? err.message : "machine export request failed",
      { code: "http" },
    );
  }

  if (!res.ok) {
    let message = `machine export failed: HTTP ${res.status}`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // keep fallback message
    }
    throw new LockKeepError(message, { code: "http", status: res.status });
  }

  return (await res.json()) as MachineEnvelope;
}

/**
 * Fetches and decrypts every secret bound to the API key's environment.
 * This is the entry point used by CI pipelines, app servers, and CLIs.
 */
export async function fetchSecrets(options: FetchOptions): Promise<SecretsSet> {
  const envelope = await fetchEnvelope(options);
  return decryptEnvelope(envelope, options.apiKey);
}

/**
 * Convenience helper: maps decrypted "env"-kind secrets to a flat
 * `KEY=value` object ready for `process.env` / `.env` output.
 */
export function toEnvVars(secrets: SecretsSet): Record<string, string> {
  const out: Record<string, string> = {};
  for (const secret of secrets.secrets) {
    out[secret.key] = secret.value;
  }
  return out;
}

/**
 * One-shot helper for the environment-variable use case:
 * fetches + decrypts, then returns `{ KEY: value }` for every secret.
 */
export async function fetchEnvVars(options: FetchOptions): Promise<Record<string, string>> {
  return toEnvVars(await fetchSecrets(options));
}