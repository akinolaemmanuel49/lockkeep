## Backend README

# LockKeep Server

Go backend for LockKeep — a zero-trust secrets manager with client-side encryption.

## What It Does

- Serves encrypted vault items without ever seeing plaintext secrets
- Manages user authentication via JWT (access tokens) and Auth0 (OAuth2)
- Enforces crypto policies: defines active KDF parameters (scrypt, Argon2id) and handles secret version stamping
- Orchestrates vault-wide operations: password verification, KDF upgrades, and bulk re-encryption coordination

## Tech Stack

- **Go 1.26+** with Gin framework
- **MongoDB** (replica set for production)
- **Auth0** for OAuth2 identity management

## Project Structure

```
├── cmd/api/          # Application entrypoint
├── internal/
│   ├── config/          # API configuration
│   ├── middleware/      # Middleware
│   ├── domain/          # Core business entities (User, VaultItem, CryptoPolicy)
│   ├── service/         # Business logic (vault, auth, crypto policy)
│   ├── handlers/        # HTTP handlers (Gin controllers)
│   ├── repository/      # MongoDB data access
│   └── dto/             # Request/response contracts
├── pkg/                 # Shared utilities
│   ├── jwt/             # JWT utility
└── go.mod
```

## Configuration

Environment variables:

| Variable | Purpose |
|----------|---------|
| `PORT` | Server listen port |
| `MONGODB_URI` | MongoDB connection string (replica set recommended) |
| `JWT_SECRET` | Symmetric key for access tokens (≥32 chars) |
| `JWT_REFRESH_SECRET` | Symmetric key for refresh tokens (independent) |
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | Auth0 application client ID |
| `AUTH0_CLIENT_SECRET` | Auth0 application client secret |
| `AUTH0_REDIRECT_URI` | OAuth2 callback URL |
| `ALLOWED_ORIGINS` | CORS allowlist |

## Running Locally

```bash
go mod download
go run ./cmd/server
```

## Docker

```bash
docker build -t lockkeep-server .
```

The server exposes port `8000` and expects to connect to MongoDB via the `MONGODB_URI` environment variable.

## Crypto Policy

The server maintains a singleton `CryptoPolicy` document that defines:
- **KDF algorithm**: `scrypt` or `argon2id`
- **Cost parameters**: tuned to your threat model
- **Version**: monotonic integer for secret lineage tracking

When secrets are updated with the `updatePolicy` flag, the server stamps them with the current policy version — clients never dictate versions directly.
