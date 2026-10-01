import { describe, expect, it, vi } from "vitest";
import {
  LockKeepError,
  WRAP_ALGORITHM,
  decryptEnvelope,
  fetchEnvVars,
  fetchEnvelope,
  fetchSecrets,
  toEnvVars,
} from "../src/index";

const utf8 = (value: string) => new TextEncoder().encode(value);
const b64 = (bytes: Uint8Array) => Buffer.from(bytes).toString("base64");

async function aesGcmEncrypt(
  key: Uint8Array,
  plaintext: Uint8Array,
  aad: Uint8Array,
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

async function deriveClientKek(apiKey: string, salt: string) {
  const material = await crypto.subtle.importKey(
    "raw",
    utf8(apiKey),
    "HKDF",
    false,
    ["deriveBits"],
  );
  return new Uint8Array(
    await crypto.subtle.deriveBits(
      {
        name: "HKDF",
        hash: "SHA-256",
        salt: utf8(salt),
        info: utf8("lockkeep:client-dek-wrap"),
      },
      material,
      32 * 8,
    ),
  );
}

const ENV_ID_HEX = "65b3f7edd9bfca00daa6e3b31";
const API_KEY = "lk_live_" + "ab".repeat(24);
const SECRETS = [
  { key: "DATABASE_URL", kind: "env", version: 1, value: "postgres://u:p@db:5432/app" },
  { key: "WRITEUP", kind: "env", version: 2, value: "driven by the docs" },
];

async function buildEnvelope() {
  const dek = crypto.getRandomValues(new Uint8Array(32));
  const kek = await deriveClientKek(API_KEY, ENV_ID_HEX);

  const wrappedDek = await aesGcmEncrypt(kek, dek, utf8(ENV_ID_HEX));

  const secrets = [];
  for (const { key, kind, version, value } of SECRETS) {
    const sealed = await aesGcmEncrypt(
      dek,
      utf8(value),
      utf8(`${ENV_ID_HEX}|${key}|${version}`),
    );
    secrets.push({ key, kind, version, ...sealed });
  }

  return {
    environmentId: ENV_ID_HEX,
    kdf: {
      algorithm: "hkdf-sha256",
      salt: ENV_ID_HEX,
      info: "lockkeep:client-dek-wrap",
      length: 32,
    },
    wrappedDek: {
      keyId: "65b3f7edd9bfca00daa6e3b32",
      algorithm: WRAP_ALGORITHM,
      ...wrappedDek,
      aad: ENV_ID_HEX,
    },
    secrets,
  };
}

describe("decryptEnvelope", () => {
  it("recovers every secret value written with the documented AAD convention", async () => {
    const envelope = await buildEnvelope();
    const result = await decryptEnvelope(envelope, API_KEY);
    expect(result.environmentId).toBe(ENV_ID_HEX);
    expect(result.secrets).toHaveLength(SECRETS.length);
    expect(result.secrets[0]).toEqual({
      key: "DATABASE_URL",
      kind: "env",
      version: 1,
      value: "postgres://u:p@db:5432/app",
    });
    expect(result.secrets[1].value).toBe("driven by the docs");
  });

  it("fails if the API key does not match", async () => {
    const envelope = await buildEnvelope();
    await expect(
      decryptEnvelope(envelope, API_KEY.slice(0, -1)),
    ).rejects.toThrow(LockKeepError);
  });

  it("fails when the kdf salt is not the environment-id string (guards the UTF-8 convention)", async () => {
    const envelope = await buildEnvelope();
    const tampered = {
      ...envelope,
      kdf: { ...envelope.kdf, salt: "00" },
    };
    await expect(decryptEnvelope(tampered, API_KEY)).rejects.toThrow(
      LockKeepError,
    );
  });

  it("rejects an unsupported kdf algorithm before doing crypto work", async () => {
    const envelope = await buildEnvelope();
    const tampered = {
      ...envelope,
      kdf: { ...envelope.kdf, algorithm: "pbkdf2-sha256" },
    };
    await expect(decryptEnvelope(tampered, API_KEY)).rejects.toThrow(
      /unsupported kdf algorithm/,
    );
  });
});

describe("fetchSecrets", () => {
  it("fetches the envelope and decrypts it end-to-end", async () => {
    const envelope = await buildEnvelope();
    const fetchMock = vi.fn(async () =>
      new Response(JSON.stringify(envelope), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchSecrets({
      apiKey: API_KEY,
      baseUrl: "https://lockkeep.example.com/api/v2/",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "https://lockkeep.example.com/api/v2/machine/secrets",
      expect.objectContaining({
        method: "GET",
        headers: { Authorization: `Bearer ${API_KEY}` },
      }),
    );
    expect(result.secrets.find((s) => s.key === "WRITEUP")?.value).toBe(
      "driven by the docs",
    );

    vi.unstubAllGlobals();
  });

  it("surfaces the server error on a failed export", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        new Response(JSON.stringify({ error: "environment not found" }), {
          status: 404,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );

    await expect(
      fetchEnvelope({ apiKey: API_KEY }),
    ).rejects.toMatchObject({ code: "http", status: 404 });

    vi.unstubAllGlobals();
  });
});

describe("helpers", () => {
  it("toEnvVars flattens secrets into KEY=value entries", async () => {
    const envelope = await buildEnvelope();
    const set = await decryptEnvelope(envelope, API_KEY);
    expect(toEnvVars(set)).toEqual({
      DATABASE_URL: "postgres://u:p@db:5432/app",
      WRITEUP: "driven by the docs",
    });
  });

  it("fetchEnvVars returns the flat map", async () => {
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
    const vars = await fetchEnvVars({ apiKey: API_KEY });
    expect(vars.DATABASE_URL).toBe("postgres://u:p@db:5432/app");
    vi.unstubAllGlobals();
  });
});