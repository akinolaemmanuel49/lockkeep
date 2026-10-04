# LockKeep — Organizations & Shared Secrets: Implementation Plan

> Draft v1 — for review before implementation. Based on the current `main` monorepo
> (backend + frontend histories merged, submodules removed) and the in-progress
> organization work on the backend.

---

## 1. Goal & Target Use Case

LockKeep currently handles **personal, single-user secrets** (zero-knowledge vault).
This plan extends it to **organizations with shared secrets**, targeting the primary
use case of **environment-variable injection**:

```
1. Operator stores an env var (e.g. DATABASE_URL, AWS keys) in an org
2. They mint a scoped API key for a machine (CI runner, app server)
3. The machine calls LockKeep with the API key
4. LockKeep authorizes + audits the call, then serves the secrets
   ENCRYPTED (server never sees the plaintext again after storage)
5. The client decrypts locally and injects into its environment
```

Design principle inherited from the existing `v2` direction
(`lockkeep-backend/V2 ARCHITECTURE.md`):

> **Personal secrets: zero-knowledge. Org/team secrets: server-managed with RBAC
> and audit logging** — extended here with a *client-decryptable* delivery mode
> so secrets never need to be read by the server to reach a machine.

Two shared-secret modes are therefore in scope:

| Mode | Who decrypts | Server reads plaintext? | Best for |
|------|-------------|------------------------|----------|
| **A. Server-managed (KMS)** | Server, then serve over TLS | Yes | Human UI, shared team vaults |
| **B. Envelope / E2EE delivery** | The authorized client only | No (after storage) | Env var injection, CI, machines |

Mode B is the new capability that makes the env-var use case safe.

---

## 2. Where We Are Today (Current State)

### What exists already (backend, `org-team` lineage)
- **Domain layer** — `User`, `Identity`, `Organization`, `Team`, `Membership`,
  `Role` / `Permission`, `AuditEvent`, `SharedSecret`, `EncryptedKey`, `MasterKeyRef`,
  `CryptoPolicy`, `VaultItem` (`lockkeep-backend/internal/domain/*.go`).
- **RBAC** — system/org/team roles with static permission maps, `RequireOrgAccess`
  and `RequirePermission` middleware; team roles nested in `Membership`
  (`internal/middleware/rbac.go`, `internal/utils/has_permission.go`).
- **Org/Team CRUD services + handlers + unit tests** — wired in `cmd/api/main.go`
  (`/api/v2/organizations/*`, `/teams/*`).
- **Auth v2** — local (bcrypt) + OAuth (Auth0), refresh tokens, JWT claims carry
  org memberships + team roles (`internal/services/auth_service.go`, `pkg/jwt`).

### What is missing / in-flight
- `SharedSecretService`, `KeyManagementService` are **empty stubs**
  (`internal/services/*_service.go` are single-line files). Shared-secret repo
  interface and domain exist; no HTTP handlers, no KMS implementation.
- `MasterKeyRef` / `EncryptedKey` schema exists; **KMS interface + implementation**
  is open (AWS KMS, HashiCorp Vault, or local dev provider).
- Personal vault flows are **commented out** in `main.go` and `vault_handler.go`
  and the worktree has an **uncommitted refactor** of `vault_service.go` /
  `ports/services.go` / `internal/dto/requests.go` (in progress).
- No **machine/service-account auth**, no **Application/Environment** model —
  required for the env-var use case.
- Frontend has no org routes yet (personal vault UI only).

---

## 3. Proposed Architecture

### 3.1 Identity model (humans + machines)

Keep `User`/`Identity` for humans (existing). Add machines as a parallel identity
type rather than bending `User`:

```
ServiceAccount
  id, org_id, name, description
  created_by, created_at, disabled_at
ApiKey                       // credential for the ServiceAccount
  _id, service_account_id, key_prefix (lk_live_xxxx)
  key_hash                    // HMAC-SHA256(key, server_secret) — never raw key
  scopes[]                    // "env:read", "secrets:read"
  expires_at, last_used_at, revoked_at
```

- **Secret check at creation**: the raw API key is shown once and stored only as
  `key_hash`. The hash is an HMAC (server-held server secret) so a DB dump alone
  cannot be used to brute-force keys via reverse lookup.
- A machine authenticates with `Authorization: Bearer lk_live_...`, gets a short
  JWT with `{ service_account_id, org_id, scopes }`, and is subject to the same
  `RequireOrgAccess` + `RequirePermission` middleware as humans.

### 3.2 Org data model for env vars

Add two entities, scoped under an existing `Organization`:

```
Application            // e.g. "web-api", "payments-worker"
  id, org_id, name, slug, description, settings
Environment           // e.g. "dev", "staging", "prod"
  id, application_id, org_id, name, slug, is_protected (prod?)
```

Then specialize shared secrets to be environment-scoped:

```
EnvSecret             // evolved SharedSecret for the machine use case
  id, org_id, application_id, environment_id
  key, version, type ("string", "file", "json", "list")
  ciphertext, nonce, key_version     // AES-256-GCM under the environment DEK
  created_by, updated_by, created_at, updated_at
  audit_events ref (read/write log per secret)
```

Compatibility: keep the existing `SharedSecret` (org/team scoped) for **Mode A /
human** flows. `EnvSecret` is the **Mode B** primitive. Both write to the same
audit log and RBAC surface.

### 3.3 Key hierarchy (Mode A + Mode B)

```
Organization
  └─ Org KEK (wrapped by KMS)            ── Mode A (server-managed)
        └─ Environment DEK (random 32B AES-256)
              └─ EnvSecret ciphertexts    (server decrypts on read)
                    ↓ on read for machine
              re-encrypt to client key    (see "conversion on read")

Org KEK (client-held, derived from org secret/passphrase — optional)
  └─ Environment DEK, wrapped to authorized clients
        └─ EnvSecret ciphertexts          ── Mode B (true E2EE)
              client unwraps DEK → decrypts locally
```

- **Mode A**: the existing `EncryptedKey`/`MasterKeyRef` model — one DEK per
  environment wrapped by a KMS-backed KEK. Server decrypts to serve the UI.
- **Mode B**: the DEK (or a per-secret key) is wrapped to the *client's* key, not
  kept server-decryptable:
  - Simple v1: DEK is **wrapped to the service account** by deriving a wrapping
    key from the client API key secret material (PBKDF2/Argon2 → KEK) at mint
    time and stored as `service_account_keys.dek_wrap`. Server only ever holds
    `wrapped_dek`; on fetch it returns `wrapped_dek` + ciphertexts; the client
    unwraps and decrypts. **Server can never read Mode B secrets.**
  - Advanced v2 (future): X25519 public-key group encryption (org members hold
    key shares; rotating a member = re-encrypting wraps). Schema designed for
    this in `EncryptedKey.Algorithm`/`KeyVersion`.

"Conversion on read" (Mode A → never-leak mode): when a Mode A secret is read by
a machine, the server can re-encrypt it under the machine's wrapping key in
memory and return the ciphertext — the server transiently decrypts KMS material
but the machine still sees only encrypted bytes. If full E2EE is desired,
prefer Mode B end to end.

### 3.4 API surface (`/api/v2`, authenticated, RBAC-guarded)

```
# Org/team (humans) — largely in progress already
GET/POST /organizations
GET/PATCH/DELETE /organizations/:orgSlug
GET/POST  /organizations/:orgSlug/members                (invite, remove)
GET/POST  /organizations/:orgSlug/teams
GET/PATCH/DELETE /organizations/:orgSlug/teams/:teamSlug

# Shared secrets — Mode A (server-managed)
GET/POST /organizations/:orgSlug/secrets
GET/PUT/DELETE /organizations/:orgSlug/secrets/:id       (audited reads)

# Service accounts (machines) — new
POST /organizations/:orgSlug/service-accounts             (returns API key once)
GET  /organizations/:orgSlug/service-accounts
PATCH/DELETE .../service-accounts/:id                     (rotate/revoke key)

# Applications & environments — new
GET/POST /organizations/:orgSlug/applications
GET/PATCH/DELETE .../applications/:appSlug
GET/POST .../applications/:appSlug/environments
GET/PATCH/DELETE .../environments/:envSlug

# Env secrets — Mode B (client-decryptable)
POST   .../applications/:appSlug/environments/:envSlug/secrets   (create/upsert by key)
GET    .../environments/:envSlug/secrets?format=env              (returns wrapped DEK + ciphertexts or plaintext env file)
GET/PUT/DELETE .../secrets/:id
POST   .../environments/:envSlug/export                          (bulk export for env injection)

# Audit
GET /audit/organizations/:orgSlug
GET /audit/environments/:envSlug        (per-env secret access trail)
```

`?format=env` renders `KEY="value"` on the client after decryption; `export`
returns a signed bundle so a CI runner can fetch once and verify integrity.

### 3.5 DB / index changes (MongoDB)

```
service_accounts   { organization_id }                       index: organization_id
api_keys           { service_account_id }                    index: key_hash (unique, partial)
apps               { organization_id, slug } unique          index: org_id
environments       { application_id, slug } unique           index: app_id
env_secrets        { environment_id, key, version } unique   index: env_id, key
audit_events       reuses existing indexes; add env_secret reads
```

All new collections use the existing `Collection[T,K]` + `UnitOfWork` infra
(`internal/infrastructure/persistence/mongo`).

### 3.6 Security requirements

- **KDF parity bug to decide**: `crypto.ts` maps `memory` to argon2 `m` (KiB) but
  scrypt `N = memory*1024` — the same policy number means different work per
  algorithm. Normalize: store per-algorithm params with explicit units.
- **Add AAD (additional authenticated data)** to AES-GCM for all secrets
  (personal + shared): bind `{user_id|env_id, key, version}` so ciphertexts can't
  be replayed across items/rows by a DB attacker with write access.
- **Version trust**: stop letting the client decide `secret.version` on update
  (`updatePolicy` query param today). Server derives version from current
  `CryptoPolicy`.
- **API keys**: store HMAC hashes only; support prefix lookup, per-key scopes,
  expiry, revocation, and rate limiting for `export`.
- **Audit**: reads of Mode B secrets log who/what/version (no plaintext);
  immutable via append-only collection + TTL retention (already designed).
- **KMS abstraction**: `KeyProvider` interface (`domain/key_management.go`)
  with three implementations: `local` (AES-KW with env-derived master key, for
  dev/tests), `aws_kms`, `hashicorp_vault`. Never log key material.

---

## 4. Implementation Phases

Each phase has an **exit criterion**; do not start the next phase without it.

### Phase 0 — Stabilize the vault foundation (prereq)
- [x] Finish the in-progress refactor: `ports/services.go` renames
  (`GetUser`/`UpdateUser`, org/team renames + orgID-based update + bulk deletes),
  reconnected personal-vault routes in `cmd/api/main.go`
  (`/api/v2/me/vault`, `/vault/items` CRUD), new `Vault`/`VaultItem`
  repositories + services + handlers, and tests (auth, org, team, user, vault).
  `go build ./...`, `go test ./...`, and `go vet ./...` are green.
- [x] Wire the authed-side vault flows: `SetVerificationHash`
  (`POST/PUT /me/vault/setup`, stores `user.vault.{verification_hash,kdf}`),
  `GetKDFParams` (`GET /me/vault/kdfparams`), `UpdateEmail`
  (`POST /me/email`, local-only + token re-issue), `UpdateAccountPassword`
  (`POST /me/password`, bcrypt re-hash), plus `VerifyVaultPassword`
  (`POST /me/vault/verify`, constant-time hash check → `{success, credentials}`).
- [x] Route-prefix unification: backend mounts profile/vault under `/me/*`
  (per arch doc), `VITE_LOCKKEEP_API_URI` moved `/api/v1` → `/api/v2`, all
  `app/lib/api/*` calls rewritten. Missing: contract tests vs running server.
- [x] Client-side crypto unit tests (`app/lib/crypto.test.ts`, vitest, 13 tests).
- [~] Decide KDF unit normalization (3.6) and CryptoPolicy versioning — open;
  backend `CryptoPolicy` service/handler still commented (repo exists), so
  `GET/POST /crypto-policy/current` 404s until Phase 2 sibling wiring.
- **Exit**: `go test ./...` green (done), frontend vault flow works end to end
  (pending backend `GET/POST /crypto-policy/current` — the unlock path calls
  `getCurrentPolicy()`).

### Phase 1 — Complete org/team RBAC (mostly scaffolded, finish it)
- Member invite/remove/role-change flows (`MembershipRepository` already has the
  methods; `buildClaims` already emits org claims).
- Audit logging middleware for org/secret reads.
- Org settings (default team role, `Require2FA` flag).
- **Exit**: org + team + membership + audit APIs covered by tests; a non-owner
  member cannot read outside their role.

### Phase 2 — Shared secrets, Mode A (server-managed)
- `KeyProvider` interface + `local` KMS dev provider.
- `EncryptedKey` provisioning per org (+ later per environment upon creation).
- Implement `SharedSecretService` + handlers + audit-on-read middleware.
- Secret versioning on write; key-version tracking (`KeyVersion` already in schema).
- **Exit**: create/read/update/delete org secrets with correct RBAC, all audited;
  demoable via curl/UI.

### Phase 3 — Env-var use case, Mode B (machines)
- `ServiceAccount` + `ApiKey` (minting, hashed keys, scopes, rotation, revocation).
- `Application` + `Environment` CRUD.
- `EnvSecret` (upsert by key): written as ciphertext under env DEK; DEK wrapped.
- `GET .../secrets?format=env` + `POST .../export`: return `wrapped_dek` +
  ciphertext bundle (never plaintext for machine clients).
- CLI/SDK (or a small fetch script) that unwraps the DEK client-side and prints
  env vars / writes a `.env` file — the exact "set API key → fetch secrets →
  decrypt locally → consume" loop.
- **Exit**: a CI runner can fetch a full environment, decrypt locally, and boot
  an app with injected env vars, with audits on every read.

### Phase 4 — Hardening & group crypto
- AAD everywhere; verifier touch-ups; rate limiting; 2FA enforcement on sensitive
  reads per org policy.
- X25519 group key wrapping (member-key-set rotation, removal = re-wrap).
- Rotate-on-read for KMS-backed keys; optional real-time audit streaming.
- Key versioning + history for org secrets (currently "schema only").

---

## 5. Open Questions (to decide before implementation)

1. Single org per user vs many (Slack model)? Current design: many (`Membership.rut.org`).
2. Should `EnvSecret` merge into `SharedSecret` (scope=environment) instead of a
   separate collection? Recommendation: yes — merge when Mode A reaches parity,
   keep the split during Phase 3 to reduce blast radius.
3. KMS provider for self-hosted (Vault) vs cloud-only? Recommend `local` default
   + AWS/Vault adapters behind the same interface.
4. API key delivery: show-once at creation + optional `invite` email? Env var
   injection means users will script creation — accept `POST` returning raw key,
   but log + require `confirm` on list.
5. Env export format: `.env` file, `export KEY=...` lines, or JSON bundle parsed by
   an SDK? Support all three behind one fetch endpoint (format param).

---

## 6. Current ZK Implementation Score (MVP)

Personal vault, scored against "correct-by-default client-side encryption" MVPs
(Bitwarden/1Password-style single-user):

| Criterion | Grade | Notes |
|-----------|-------|-------|
| Key derivation & splitting | **8/10** | argon2id/scrypt, 64B output → 32B AES key + 32B verifier; noble-hashes + WebCrypto. KDF unit mismatch across algorithms (scrypt N vs argon m). |
| Encryption primitives | **8/10** | AES-256-GCM, 12B random nonce per item, tag split handled correctly. Missing AAD binding metadata/version. |
| Client key hygiene | **9/10** | Non-extractable CryptoKey, memory-only, 5-min auto-lock, derived buffers zeroed after use. |
| Server trust boundary | **8/10** | Server stores only ciphertext/IV/tag + hashed verifier; never sees plaintext; query scoped by user+tenant. |
| Crypto-policy migration | **6/10** | Detection + transparent re-encryption exist, but `secret.version` is client-driven (`updatePolicy`) — server should own versioning. |
| Verification | **7/10** | Verifier is uncorrelated SHA-256 of KDF tail (fine); but salts/params exposed to any logged-in user; acceptable for MVP, improve with per-user salt isolation and constant-time compare (already `subtle`). |
| Testing | **5/10** | Backend auth/tests exist; client crypto + vault service untested; migration path untested. |
| **Overall** | **7.5/10** | Solid MVP foundations; fix KDF-units, add AAD + server-owned versioning, and add crypto tests before building org features on top. |