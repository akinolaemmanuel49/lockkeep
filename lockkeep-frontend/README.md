## Frontend README

# LockKeep UI

React frontend for LockKeep — a zero-trust secrets manager where all encryption happens in the browser.

## What It Does

- Encrypts secrets client-side before they ever reach the server
- Manages vault lifecycle: unlock, lock, password change, KDF migration
- Caches decrypted secrets in-memory with automatic expiry (5-minute vault sessions)
- Transparently re-encrypts all vault items when crypto policies change

## Tech Stack

- **React Router v7** with SSR/SPA hybrid
- **Vite** for builds
- **Auth0 SPA SDK** for OAuth2 authentication

## Project Structure

```
├── app/                 # React Router routes and pages
│   ├── components/      # Reusable UI components
│   ├── hooks/           
│   ├── providers/       
│   ├── routes/          
│   ├── types/           # TypeScript definitions
│   ├── utils/           
│   ├── lib/ 
│   │   └── api/         # HTTP clients     
│   ├── app.css  
│   ├── config.ts  
│   ├── root.ts
│   └── routes.ts
└── vite.config.ts
```

## Configuration

Environment variables:

| Variable | Purpose |
|----------|---------|
| `VITE_LOCKKEEP_API_URI` | Backend base URL |
| `VITE_AUTH0_REDIRECT_URI` | OAuth2 callback URL |
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | Auth0 SPA client ID |

## Running Locally

```bash
npm install
npm run dev
```

## Docker

```bash
docker build -t lockkeep-ui .
```

The frontend serves on port `3000` and expects the backend at `VITE_LOCKKEEP_API_URI`.

## Record Types

The vault accepts six record types, each rendered with a type-specific form and
card layout via `app/lib/recordTypes.tsx`:

| Type | Form fields (metadata) | Secret field |
|------|------------------------|--------------|
| `login` | site URL, identifier | password (with generator) |
| `environment` | — | value |
| `ssh_key` | username | private key (multiline) |
| `secure_note` | — | note (multiline) |
| `payment_card` | card holder, expiration | card number (masked as `•••• •••• •••• ••••`, formatted on reveal) |
| `api_key` | issuer, scopes | key value |

Only the encrypted `secret` field should hold sensitive data — `metadata` is
stored in plaintext by design.

## Vault Architecture
(scrypt/Argon2id)
```
┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│    Vault    │     │     KDF         │     │ Encryption  │
│  Password   │──── │(scrypt/Argon2id)│──── │    Key      │
└─────────────┘     └─────────────────┘     └──────┬──────┘
                                                   │
                            ┌──────────────────────┼─────────────────────┐
                            │                      │                     │                                          
                        ┌─────────┐           ┌─────────┐           ┌─────────┐
                        │ Secret 1│           │ Secret 2│           │ Secret N│
                        │(AES-GCM)│           │(AES-GCM)│           │(AES-GCM)│
                        └─────────┘           └─────────┘           └─────────┘
```

- **Vault unlock**: Derives key from password + stored KDF params, verifies hash, decrypts all secrets into memory cache
- **KDF migration**: On unlock or password change, detects policy drift, re-derives key with new params, re-encrypts everything server-side
- **Vault lock**: Wipes key and cache from memory
- **Seamless unlock**: If the token expires (or the vault locks) while a record form is
  open, saving re-prompts for the master password and completes automatically —
  typed input is preserved. Expired decrypt/reveal/copy actions retry after unlocking.