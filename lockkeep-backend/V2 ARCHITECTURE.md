## LockKeep v2 Architecture

### Core Principle

**Personal secrets: zero-knowledge. Org/team secrets: server-managed with RBAC and audit logging.**

---

## Domain Model

| Entity         | Responsibility                              | Storage           |
| -------------- | ------------------------------------------- | ----------------- |
| `User`         | Profile, personal vault metadata            | `users`           |
| `Identity`     | Authentication credentials (local, OAuth)   | `identities`      |
| `Organization` | Top-level grouping, settings                | `organizations`   |
| `Team`         | Sub-grouping within org                     | `teams`           |
| `Membership`   | User-org-team-role junction                 | `memberships`     |
| `Role`         | RBAC definitions (system, org, team scopes) | `roles` (seeded)  |
| `VaultItem`    | Personal secrets — zero-knowledge           | `vault_items`     |
| `SharedSecret` | Org/team secrets — server-managed           | `shared_secrets`  |
| `EncryptedKey` | Org/team encryption keys (wrapped by KMS)   | `encrypted_keys`  |
| `MasterKeyRef` | External KMS reference                      | `master_key_refs` |
| `AuditEvent`   | Immutable access log                        | `audit_events`    |
| `CryptoPolicy` | Platform KDF parameters                     | `crypto_policies` |

---

## Security Model Split

| Aspect               | Personal Vault                        | Org/Team Secrets                    |
| -------------------- | ------------------------------------- | ----------------------------------- |
| **Encryption**       | Client-side (user-derived master key) | Server-side (KMS-backed)            |
| **Key location**     | Never leaves client                   | Stored encrypted, decrypted via KMS |
| **Server can read?** | **No**                                | **Yes** — server decrypts to serve  |
| **Recovery**         | Impossible if password lost           | Org admin can rotate, re-encrypt    |
| **Audit**            | Client-reported (optional)            | Server-enforced, immutable          |
| **Threat model**     | Server compromise, insider            | External attacker, accidental leak  |

---

## Directory Structure

```
internal/
├── domain/                    # All structs, enums, constants
│   ├── user.go
│   ├── identity.go
│   ├── organization.go
│   ├── team.go
│   ├── membership.go
│   ├── role.go
│   ├── vault.go               # VaultItem, Secret, VaultMetadata, KDFParams
│   ├── shared_secret.go
│   ├── key_management.go      # EncryptedKey, MasterKeyRef
│   ├── audit.go
│   └── crypto_policy.go
│
├── ports/                     # Repository interfaces (dependency inversion)
│   └── repository.go          # All repository contracts
│
├── repository/                # Concrete repository implementations
│   ├── user.go
│   ├── identity.go
│   ├── organization.go
│   ├── team.go
│   ├── membership.go
│   ├── shared_secret.go
│   ├── audit.go
│   ├── vault_item.go
│   └── crypto_policy.go
│
├── service/                   # Business logic
│   ├── auth_service.go
│   ├── user_service.go
│   ├── org_service.go
│   ├── team_service.go
│   ├── rbac_service.go        # Permission evaluation
│   ├── shared_secret_service.go
│   ├── vault_service.go         # Personal vault operations
│   ├── audit_service.go
│   └── crypto_policy_service.go
│
├── handlers/                  # HTTP adapters
│   ├── auth_handler.go
│   ├── user_handler.go
│   ├── org_handler.go
│   ├── team_handler.go
│   ├── shared_secret_handler.go
│   ├── vault_handler.go
│   ├── audit_handler.go
│   └── crypto_policy_handler.go
│
├── middleware/
│   ├── auth.go                # JWT validation
│   ├── rbac.go                # Permission checks
│   ├── audit.go               # Auto-log shared secret access
│   └── tenant.go              # Org/team isolation
│
└── infrastructure/
    └── persistence/
        └── mongo/
            ├── collection.go      # Collection[T, K] interface
            ├── collection_impl.go # mongoCollection[T, K] implementation
            └── transaction.go     # UnitOfWork / session wrapper
```

---

## Repository Contracts (`ports/repository.go`)

| Interface                | Methods                                                                                                     |
| ------------------------ | ----------------------------------------------------------------------------------------------------------- |
| `UserRepository`         | `Create`, `FindByEmail`, `EmailExists`, `FindByID`, `UpdateVaultMetadata`, `UpdateEmail`                    |
| `IdentityRepository`     | `Create`, `FindByUserID`, `FindByOAuth`, `UpdatePassword`, `RecordLogin`                                    |
| `OrganizationRepository` | `Create`, `FindByID`, `FindBySlug`, `FindByMember`, `Update`, `Delete`                                      |
| `TeamRepository`         | `Create`, `FindByID`, `FindBySlug`, `FindByOrganization`, `Update`, `Delete`                                |
| `MembershipRepository`   | `Create`, `FindByUserAndOrg`, `FindByUser`, `FindByOrganization`, `UpdateRole`, `UpdateTeamRoles`, `Delete` |
| `SharedSecretRepository` | `Create`, `FindByID`, `FindByOrganization`, `Update`, `Delete`                                              |
| `AuditRepository`        | `Log`, `FindByResource`, `FindByUser`, `FindByOrganization`                                                 |
| `VaultItemRepository`    | `Create`, `FindByUser`, `FindByID`, `Update`, `Delete`                                                      |
| `CryptoPolicyRepository` | `GetCurrent`, `SetCurrent`                                                                                  |

---

## Middleware Stack

| Middleware                | Applies To                | Purpose                                      |
| ------------------------- | ------------------------- | -------------------------------------------- |
| `Auth`                    | All protected routes      | Validate JWT, set `userID` in context        |
| `RequireOrgAccess`        | Org-scoped routes         | Verify membership, set `orgID`, `membership` |
| `RequirePermission(perm)` | Specific routes           | Check role permissions                       |
| `AuditAccess`             | Shared secret read routes | Auto-log `AuditEvent` before response        |

---

## Route Structure

```
/api/v2/
├── POST /auth/register          → Create User + Identity
├── POST /auth/login             → Validate Identity, issue JWT
├── POST /auth/oauth             → OAuth flow, create/link Identity
├── POST /auth/refresh           → Refresh JWT
│
├── GET    /me                   → Get own User profile
├── PATCH  /me                   → Update profile
│
├── GET    /me/vault             → Personal vault
├── POST   /me/vault/items       → Create personal VaultItem
├── GET    /me/vault/items       → List personal VaultItems
├── GET    /me/vault/items/:id   → Get personal VaultItem
├── PUT    /me/vault/items/:id   → Update personal VaultItem
├── DELETE /me/vault/items/:id   → Delete personal VaultItem
│
├── GET    /organizations                    → List my orgs
├── POST   /organizations                    → Create org (become owner)
├── GET    /organizations/:orgSlug           → Get org
├── PATCH  /organizations/:orgSlug           → Update org (admin+)
├── DELETE /organizations/:orgSlug           → Delete org (owner)
│
├── GET    /organizations/:orgSlug/members              → List members
├── POST   /organizations/:orgSlug/members              → Invite member
├── DELETE /organizations/:orgSlug/members/:userID      → Remove member
├── PATCH  /organizations/:orgSlug/members/:userID/role   → Change org role
│
├── GET    /organizations/:orgSlug/teams                  → List teams
├── POST   /organizations/:orgSlug/teams                  → Create team
├── GET    /organizations/:orgSlug/teams/:teamSlug        → Get team
├── PATCH  /organizations/:orgSlug/teams/:teamSlug        → Update team
├── DELETE /organizations/:orgSlug/teams/:teamSlug        → Delete team
│
├── GET    /organizations/:orgSlug/teams/:teamSlug/members           → List team members
├── POST   /organizations/:orgSlug/teams/:teamSlug/members             → Add to team
├── DELETE /organizations/:orgSlug/teams/:teamSlug/members/:userID   → Remove from team
│
├── GET    /organizations/:orgSlug/secrets               → Org-wide secrets
├── POST   /organizations/:orgSlug/secrets               → Create org secret
│
├── GET    /organizations/:orgSlug/teams/:teamSlug/secrets             → Team secrets
├── POST   /organizations/:orgSlug/teams/:teamSlug/secrets             → Create team secret
├── GET    /organizations/:orgSlug/teams/:teamSlug/secrets/:id         → Get secret (audited)
├── PUT    /organizations/:orgSlug/teams/:teamSlug/secrets/:id         → Update secret
├── DELETE /organizations/:orgSlug/teams/:teamSlug/secrets/:id         → Delete secret
│
├── GET    /audit/organization/:orgSlug                  → Org audit log (admin+)
├── GET    /audit/user                                   → My audit log
│
└── GET    /crypto-policy/current                        → Current policy (public)
└── POST   /crypto-policy/current                        → Set policy (system admin only)
```

---

## Key Architectural Decisions

| Decision                                              | Why                                                                            |
| ----------------------------------------------------- | ------------------------------------------------------------------------------ |
| **User and Identity separate**                        | One user can have multiple auth methods; auth concerns don't leak into profile |
| **No `TenantID` on User**                             | Users exist independently of orgs; Slack model                                 |
| **No `UserType` enum**                                | Replaced by `Role` system with scopes (system, org, team)                      |
| **Membership as junction**                            | Clean many-to-many; one doc per user-org pair                                  |
| **Team roles nested in Membership**                   | Avoids separate collection; index `team_roles.team_id`                         |
| **Slug-based URLs**                                   | Human-readable, cacheable: `/orgs/acme/teams/engineering`                      |
| **Path versioning in one binary**                     | Simpler ops; `/api/v2/` prefix                                                 |
| **Collection wrapper**                                | Business language over MongoDB; swappable backend                              |
| **Audit auto-logged on secret access**                | Immutable, server-enforced; not client-reported                                |
| **CryptoPolicy public, mutable only by system admin** | Transparency doesn't hurt security                                             |

---

## Data Flow: Personal Vault (Zero-Knowledge)

```
Client derives master key from password
        │
        ▼
Client encrypts VaultItem with master key
        │
        ▼
POST /me/vault/items
        │
        ▼
Server stores ciphertext, IV, tag only
        │
        ▼
Server never sees plaintext, cannot decrypt
```

## Data Flow: Shared Secret (Server-Managed)

```
User requests secret via GET /orgs/:org/teams/:team/secrets/:id
        │
        ▼
Auth middleware validates JWT
        │
        ▼
RequireOrgAccess verifies membership
        │
        ▼
RequirePermission checks role
        │
        ▼
Audit middleware logs access event
        │
        ▼
Fetch EncryptedKey for org/team
        │
        ▼
KMS decrypts EncryptedKey → plaintext key
        │
        ▼
Decrypt SharedSecret with plaintext key
        │
        ▼
Return plaintext over TLS
        │
        ▼
Key evicted from memory
```

---

## Indexes

| Collection       | Index                                           | Purpose                     |
| ---------------- | ----------------------------------------------- | --------------------------- |
| `users`          | `email` unique                                  | Login lookup                |
| `users`          | `username` unique                               | Profile uniqueness          |
| `identities`     | `user_id`                                       | Find user's auth methods    |
| `identities`     | `auth_method + auth_provider_id` partial unique | OAuth deduplication         |
| `memberships`    | `user_id + organization_id` unique              | One membership per user-org |
| `memberships`    | `organization_id`                               | List org members            |
| `memberships`    | `user_id`                                       | List user's orgs            |
| `memberships`    | `team_roles.team_id`                            | Find team members           |
| `organizations`  | `slug` unique                                   | URL lookup                  |
| `teams`          | `organization_id + slug` unique                 | Scoped lookup               |
| `shared_secrets` | `organization_id + team_id`                     | List org/team secrets       |
| `audit_events`   | `resource_id + timestamp` desc                  | Audit queries               |
| `audit_events`   | `user_id + timestamp` desc                      | User history                |
| `audit_events`   | `organization_id + timestamp` desc              | Org audit                   |
| `audit_events`   | `timestamp` desc TTL 90d                        | Auto-expiration             |

---

## Open Questions / Future Work

| Topic                                             | Status                                          |
| ------------------------------------------------- | ----------------------------------------------- |
| KMS integration (AWS KMS, HashiCorp Vault, local) | TBD — interface defined, implementation pending |
| Key rotation for org/team secrets                 | Schema supports `KeyVersion`, logic pending     |
| Client-side group encryption (advanced)           | Schema designed for, not implemented            |
| Real-time audit streaming                         | Out of scope for MVP                            |
| Secret versioning / history                       | Out of scope for MVP                            |

---

Want me to expand any section — the RBAC evaluation flow, the KMS abstraction, or the audit middleware design?
