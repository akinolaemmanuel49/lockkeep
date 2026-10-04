# lockkeep-example

Minimal application that reads one LockKeep environment and exposes its
secrets as **plain environment variables** — your service code never imports
LockKeep.

## How it works

1. The `run` wrapper calls `fetchEnvVars()` — a **single API call**:
   `GET /api/v2/machine/secrets` hands back an encrypted envelope.
2. The SDK derives the wrapping key (HKDF-SHA256), unwraps the data key
   (AES-256-GCM), and decrypts every secret locally.
3. The vars are injected into `process.env` before your command starts.
4. Your app just reads `process.env.DATABASE_URL` like any other env var.

`npm run demo` runs `dist/app.js` behind the wrapper — that tiny app has zero
LockKeep imports and simply prints whether each secret is present.

## Requirements

- Node.js ≥ 18 (built against the workspace SDK).
- An env-scoped LockKeep API key with `secrets:read`
  (`lk_live_…`), minted in the console under **Service accounts**.

## Run locally

```shell
cd ../lockkeep-sdk && npm run build && cd ../lockkeep-example
npm install

$env:LOCKKEEP_API_KEY = "lk_live_…"          # PowerShell
npm run demo                                   # inject + run dist/app.js

# or run any command with the env injected
npm start -- run -- node dist/app.js --values

# preview modes (no injection)
npm start                                      # masked keys + kinds, 1 API call
npm start -- --env                            # KEY=value lines
npm start -- --assert                         # exit 3 if no secrets returned
```

`LOCKKEEP_API_URI` overrides the API base URL
(default `http://localhost:8000/api/v2`). Values stay masked in CI logs.

## Run in CI

See `.github/workflows/ci-secrets.yml`. It:

1. checks out the repo,
2. builds the SDK,
3. installs this example (its `file:` dependency triggers the SDK's
   `prepare` build automatically),
4. runs the app through the injector with `LOCKKEEP_API_KEY` /
   `LOCKKEEP_API_URI` supplied from GitHub Actions secrets — the job fails if
   the fetch errors or the app finds no secrets.

Wire it up by creating a service account and minting a `secrets:read` key in
the LockKeep console, then setting the two repository secrets:

| Secret | Value |
| --- | --- |
| `LOCKKEEP_API_KEY` | raw `lk_live_…` key |
| `LOCKKEEP_API_URI` | `https://<host>/api/v2` |