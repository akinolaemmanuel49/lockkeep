# LockKeep

**Zero Trust Secrets Manager**

LockKeep is a self-hosted, zero-knowledge secrets manager built with security-first principles. It encrypts all vault items client-side, supports modern KDF algorithms (scrypt, Argon2id), and automatically migrates secrets to match your organization's current crypto policy.

---

## Features

- **Zero-Trust Architecture** — All secrets are encrypted client-side before reaching the server. The server never sees plaintext.
- **Auto Secret Migration** — Secrets are automatically re-encrypted to match the current crypto policy on vault unlock and password change, with transparent version tracking.
- **Modern KDF Support** — Pluggable key derivation: `scrypt` and `argon2id`.
- **OAuth2 Authentication** — Managed via Auth0 with JWT access tokens (session storage) and refresh tokens (HTTP-only cookies).
- **Production-Ready Infrastructure** — Docker containerization, Traefik load balancing, and MongoDB replica sets for high availability.

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go (Gin framework) |
| Frontend | React Router v7 + Vite |
| Database | MongoDB (3-node replica set: 1 primary, 2 secondaries) |
| Load Balancer / Router | Traefik v3 |
| Auth | Auth0 (OAuth2) |
| Containerization | Docker & Docker Compose |

---

## Architecture

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│   Traefik       │──────│  lockkeep-ui    │      │ lockkeep-server │
│   (:80, :8080)  │      │   (:3000)       │      │    (:8000)      │
└─────────────────┘      └─────────────────┘      └─────────────────┘
        │                                                    │
        └────────────────────────────────────────────────────┘
                              │
                    ┌─────────────────┐
                    │  MongoDB RS     │
                    │  (3 replicas)   │
                    └─────────────────┘
```

- **Traefik** routes traffic by hostname:
  - `lockkeep.localhost` → frontend
  - `api.lockkeep.localhost` → backend
- **Scaling**: Use `docker compose up --scale server=N` to spin up multiple backend instances behind Traefik's load balancer.
- **MongoDB**: Runs as a 3-node replica set (`rs0`) for fault tolerance and transaction support.

---

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Auth0 application credentials (see [Auth0 Setup](#auth0-setup))

### 1. Create the Docker Network

Traefik and all services communicate over an external Docker network:

```bash
docker network create lockkeep-network
```

### 2. Configure Environment Variables

**Backend** (`.env` in server directory):

```env
PORT=8000
MONGODB_URI=mongodb://root:password@mongo1:27017,mongo2:27018,mongo3:27019/?replicaSet=rs0&authSource=admin
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000,http://lockkeep.localhost
JWT_SECRET=your-256-bit-secret-key-here-min-32-chars
JWT_REFRESH_SECRET=your-refresh-secret-different-from-above
AUTH0_DOMAIN=your-auth0-domain.auth0.com
AUTH0_CLIENT_ID=your-auth0-client-id
AUTH0_CLIENT_SECRET=your-auth0-client-secret
AUTH0_REDIRECT_URI=http://localhost:5173/callback
```

**Frontend** (`.env` in frontend directory):

```env
AUTH0_DOMAIN=your-auth0-domain.auth0.com
AUTH0_CLIENT_ID=your-auth0-client-id
VITE_AUTH0_REDIRECT_URI=http://localhost:5173/callback
VITE_LOCKKEEP_API_URI=http://api.lockkeep.localhost/api/v1
```

> **Security Note:** `JWT_SECRET` and `JWT_REFRESH_SECRET` must be at least 32 characters and cryptographically independent.

### 3. Initialize the MongoDB Replica Set

Start the database layer first:

```bash
docker compose -f docker-compose.db.yml up -d
```

Then initialize the replica set (one-time setup):

```bash
docker exec -it mongo1 mongosh --eval "
rs.initiate({
  _id: 'rs0',
  members: [
    { _id: 0, host: 'mongo1:27017' },
    { _id: 1, host: 'mongo2:27017' },
    { _id: 2, host: 'mongo3:27017' }
  ]
})
"
```

### 4. Start the Application

**Backend** (with optional scaling):

```bash
# Build image
docker build . -t lockkeep-server:latest
# Single instance
docker compose -f docker-compose.server.yml up -d

# Or scale horizontally
docker compose -f docker-compose.server.yml up -d --scale server=3
```

**Frontend:**

```bash
# Build image
docker build . -t lockkeep-ui:latest

docker compose -f docker-compose.frontend.yml up -d
```

### 5. Access the Application

| Service | URL |
|---------|-----|
| LockKeep UI | http://lockkeep.localhost |
| API | http://api.lockkeep.localhost |
| Traefik Dashboard | http://localhost:8080 |
| Mongo Express | http://localhost:8081 |

> Add `127.0.0.1 lockkeep.localhost api.lockkeep.localhost` to your `/etc/hosts` file if these domains don't resolve.

---

## Auth0 Setup

1. Create a new **Single Page Application** in your Auth0 dashboard.
2. Configure **Allowed Callback URLs**: `http://localhost:5173/callback`
3. Configure **Allowed Logout URLs**: `http://lockkeep.localhost`
4. Configure **Allowed Web Origins**: `http://lockkeep.localhost`
5. Copy the **Domain**, **Client ID**, and **Client Secret** into your `.env` files.

---

## Docker Compose Reference

### Database (`docker-compose.db.yml`)

```yaml
services:
  traefik:
    image: traefik:v3.7
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    networks:
      - lockkeep-network

  mongo1:
    image: mongo
    command:
      - sh
      - -c
      - |
        cp /tmp/mongo-keyfile /etc/mongo-keyfile &&
        chmod 400 /etc/mongo-keyfile &&
        exec mongod --replSet rs0 --bind_ip_all --keyFile /etc/mongo-keyfile
    restart: always
    healthcheck:
      test: [ "CMD", "mongosh", "--eval", "db.adminCommand('ping')" ]
      interval: 5s
      timeout: 5s
      retries: 10
    ports:
      - 27017:27017
    networks:
      - lockkeep-network
    volumes:
      - ./secrets/mongo-keyfile:/tmp/mongo-keyfile:ro
      - mongo1:/data/db

  mongo2:
    image: mongo
    command:
      - sh
      - -c
      - |
        cp /tmp/mongo-keyfile /etc/mongo-keyfile &&
        chmod 400 /etc/mongo-keyfile &&
        exec mongod --replSet rs0 --bind_ip_all --keyFile /etc/mongo-keyfile
    restart: always
    healthcheck:
      test: [ "CMD", "mongosh", "--eval", "db.adminCommand('ping')" ]
      interval: 5s
      timeout: 5s
      retries: 10
    ports:
      - 27018:27017
    networks:
      - lockkeep-network
    volumes:
      - ./secrets/mongo-keyfile:/tmp/mongo-keyfile:ro
      - mongo2:/data/db

  mongo3:
    image: mongo
    command:
      - sh
      - -c
      - |
        cp /tmp/mongo-keyfile /etc/mongo-keyfile &&
        chmod 400 /etc/mongo-keyfile &&
        exec mongod --replSet rs0 --bind_ip_all --keyFile /etc/mongo-keyfile
    restart: always
    healthcheck:
      test: [ "CMD", "mongosh", "--eval", "db.adminCommand('ping')" ]
      interval: 5s
      timeout: 5s
      retries: 10
    ports:
      - 27019:27017
    networks:
      - lockkeep-network
    volumes:
      - ./secrets/mongo-keyfile:/tmp/mongo-keyfile:ro
      - mongo3:/data/db

  mongo-express:
    image: mongo-express
    restart: always
    ports:
      - 8081:8081
    environment:
      ME_CONFIG_MONGODB_URL: mongodb://root:password@mongo1:27017,mongo2:27017,mongo3:27017/admin?replicaSet=rs0
      ME_CONFIG_BASICAUTH_ENABLED: true
      ME_CONFIG_BASICAUTH_USERNAME: mongoexpressuser
      ME_CONFIG_BASICAUTH_PASSWORD: mongoexpresspass
    networks:
      - lockkeep-network

networks:
  lockkeep-network:
    external: true

volumes:
  mongo1:
  mongo2:
  mongo3:
```

> **Note:** Generate a MongoDB keyfile before starting: `openssl rand -base64 756 > secrets/mongo-keyfile && chmod 400 secrets/mongo-keyfile`

### Server (`docker-compose.server.yml`)

```yaml
services:
  server:
    image: lockkeep-server
    labels:
      - "traefik.http.routers.lockkeep-server.rule=Host(`api.lockkeep.localhost`)"
      - "traefik.http.services.lockkeep-server.loadbalancer.server.port=8000"
    ports:
      - 8000:8000
    depends_on:
      mongo1:
        condition: service_healthy
      mongo2:
        condition: service_healthy
      mongo3:
        condition: service_healthy
    env_file:
      - .env
    networks:
      - lockkeep-network

networks:
  lockkeep-network:
    external: true
```

### Frontend (`docker-compose.frontend.yml`)

```yaml
services:
  frontend:
    image: lockkeep-ui
    labels:
      - "traefik.http.routers.lockkeep-ui.rule=Host(`lockkeep.localhost`)"
      - "traefik.http.services.lockkeep-ui.loadbalancer.server.port=3000"
    env_file:
      - .env
    networks:
      - lockkeep-network

networks:
  lockkeep-network:
    external: true
```

---

## Crypto Policy & Secret Versioning

LockKeep's crypto policy defines the active KDF parameters (algorithm, cost factors, salt length). When a policy changes:

1. **Detection**: On unlock, the client compares the user's stored KDF params with the current policy.
2. **Auto-Migration**: If they differ, all vault items are transparently decrypted with the old key and re-encrypted with a key derived under the new policy.
3. **Version Tracking**: Each secret is tagged with the policy version used for its encryption, enabling auditability and future migration paths.

Supported KDFs:
- `scrypt` (N, r, p configurable)
- `argon2id` (memory, iterations, parallelism configurable)

---

## Development

### Backend

```bash
cd server
go mod download
go run ./cmd/api
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```
