# @lockkeep/sdk

TypeScript client for **LockKeep machine exports**. Fetch an environment's
encrypted envelope with a `lk_live_…` API key and decrypt it locally.

The machine endpoint only ever returns ciphertext — the server derives a
per-key wrapping key (HKDF-SHA256) and wraps each environment's data key
(AES-256-GCM), and seals every secret under that data key. Only a client in
possession of the raw API key can derive the wrapping key and read the values.
This SDK implements that unwrap/decrypt pipeline with zero runtime
dependencies (WebCrypto + `fetch`), so it runs on Node.js ≥ 18 and in
browsers.

## Install

```shell
npm install /path/to/lockkeep-sdk   # or publish it and use @lockkeep/sdk
```

## Quick start

```ts
import { fetchEnvVars } from "@lockkeep/sdk";

const vars = await fetchEnvVars({
  apiKey: "lk_live_…",
  baseUrl: "https://lockkeep.example.com/api/v2",
});

process.env = { ...process.env, ...vars };
```

## API

| Export | Description |
| --- | --- |
| `fetchSecrets({ apiKey, baseUrl?, signal? })` | Fetch + decrypt every secret → `{ environmentId, secrets: [{ key, kind, version, value }] }` |
| `fetchEnvVars(options)` | `fetchSecrets` mapped to a flat `{ KEY: value }` object |
| `toEnvVars(secretsSet)` | Map an already-decrypted set to `{ KEY: value }` |
| `fetchEnvelope(options)` | Fetch the raw ciphertext envelope only |
| `decryptEnvelope(envelope, apiKey)` | Unwrap + decrypt a raw envelope |
| `unwrapEnvelope(envelope, apiKey)` | Derive the KEK and return the unwrapped DEK as a `CryptoKey` |
| `decryptSecret(dek, entry, environmentId)` | Decrypt a single entry with an unwrapped DEK |
| `LockKeepError` | Enriched error (`code`, optional `status`) |
| `HKDF_INFO` / `WRAP_ALGORITHM` | The wire constants |

## Key facts a client must honor

- **Auth:** `Authorization: Bearer <raw-api-key>` on `GET /api/v2/machine/secrets`.
  The key is bound to exactly one environment at mint time — nothing else selects it.
- **KDF:** `HKDF-SHA256(ikm = raw key UTF-8, salt = environmentId hex STRING
  as UTF-8 bytes, info = "lockkeep:client-dek-wrap", length = 32)`.
- **Wrap:** `AES-256-GCM(dek, aad = environmentId hex STRING as UTF-8 bytes)`.
- **Secrets:** `AES-256-GCM(value, aad = "${environmentId}|${key}|${version}")`;
  ciphertext/nonce are standard base64.
- **Never hex-decode `kdf.salt` or `wrappedDek.aad`** — they are the hex string's
  characters, not its bytes. This trips up hex-happy implementations (and is
  covered by a regression test).

## Development

```shell
npm install
npm run typecheck
npm test
npm run build
```

## Pairing with the web console

The console lets an operator create a service account, mint an environment-key
(`scopes: ["secrets:read"]`), and instantly verify the same unwrap pipeline in
the browser — the exact operations this SDK automates in CI.