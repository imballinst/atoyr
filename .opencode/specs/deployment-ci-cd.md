# Deployment & CI/CD

## Rationale

The production VM is a 4GB/2vCPU box that also runs Coolify. Building container images locally (the Coolify-native way) risks OOM kills because `npm install`/webpack/Go compilation spikes memory while app containers are already running. Coolify itself recommends ~4GB/2vCPU just for its own operation, so builds on-server are unsafe.

What's ruled out:

- **Local builds in Coolify** — OOM risk on this box.
- **Self-hosted Harbor** — too heavy (core, DB, Redis, registry, portal, jobservice) for the remaining headroom.
- **Docker Hub private repos** — works but requires a paid Docker Hub Pro plan; unnecessary cost since the code is on GitHub.

## Chosen approach: GitHub Actions → GHCR (private) → Coolify pulls

```
┌──────────────┐     push ghcr.io     ┌──────────────┐    docker pull    ┌─────────┐
│ GitHub       │ ──────────────────→  │ GHCR          │ ───────────────→ │ Coolify │
│ Actions      │                      │ (private)     │                  │ (VM)    │
│ (build+push) │                      │               │                  │         │
└──────────────┘                      └──────────────┘                  └─────────┘
```

### Why it works

- **Private images are free** on GHCR. Default visibility is private; when linked to a private repo, it inherits the repo's visibility automatically. No Docker Hub subscription needed.
- **No extra secrets for push**. Inside the Actions run, `GITHUB_TOKEN` authenticates the push — no manually-managed registry credential.
- **Zero build load on the VM**. GitHub runners handle compilation; Coolify only runs `docker pull`.
- **No extra containers on the VM**. Unlike Harbor, nothing to host locally.

### The one wrinkle: pulling private images

Coolify needs a GitHub Personal Access Token (PAT) with `read:packages` scope (add `repo` scope too if the package inherits visibility from a private repo — GitHub requires it). Unlike the push side (which uses `GITHUB_TOKEN`), the pull side must use a PAT because `GITHUB_TOKEN` only exists inside the Actions runtime.

On the VM:

```bash
echo $PAT | docker login ghcr.io -u <github-username> --password-stdin
```

This logs the local Docker daemon into GHCR. Coolify relies on Docker's own credential store — it doesn't manage registry passwords itself, so this one-time `docker login` makes subsequent pulls work.

In Coolify:
- Use a "prebuilt image" deployment pointing at `ghcr.io/atoyr/app:latest`.

## Implementation

### 1. Single Dockerfile at `Dockerfile` (project root)

Multi-stage build that compiles both the client SPA and the Go server, then packages the Go binary together with the client's static assets.

### `.dockerignore`

At the project root, to keep the build context lean and avoid sending `node_modules` to the Docker daemon:

```
node_modules
.git
.gitignore
.yarn/cache
packages/client/build
packages/server/bin
.env.local
README.md
```

### `Dockerfile`

```dockerfile
# ---- Client build ----
FROM node:24-alpine AS client-build

WORKDIR /app

COPY package.json yarn.lock ./
COPY packages/client/package.json packages/client/package.json
COPY .yarnrc.yml ./

RUN yarn workspaces focus client

COPY packages/client/ packages/client/

RUN yarn workspace client build

# ---- Server build ----
FROM golang:1.25-alpine AS server-build

WORKDIR /app

COPY packages/server/go.mod packages/server/go.sum ./
RUN go mod download

COPY packages/server/ .

RUN CGO_ENABLED=0 go build -o /app/server ./cmd/main.go

# ---- Production ----
FROM nginx:1.27-alpine

# Go server
COPY --from=server-build /app/server /app/server/server
COPY packages/server/data/words.json /app/server/data/words.json
COPY packages/server/web/static/ /app/server/web/static/

# Client SPA
COPY --from=client-build /app/packages/client/build/client/ /usr/share/nginx/html/

# Nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Entrypoint to run Go server and nginx
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/docker-entrypoint.sh"]
```

nginx acts as reverse proxy: it serves SPA static files directly and proxies `/api/` and `/dashboard` to the Go server on port 3000 (running in the same container). A small entrypoint script starts the Go server in the background, then launches nginx in the foreground. The Go server needs no changes for static file serving; it stays pure API + dashboard backend.

An entrypoint script at `docker-entrypoint.sh` starts the Go server and then nginx:

```bash
#!/bin/sh
set -e

/app/server/server &

exec nginx -g "daemon off;"
```

### `nginx.conf`

At the project root:

```nginx
server {
    listen 80;
    server_name _;

    location /api/ {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /dashboard {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }
}
```

### 2. `Makefile`

For local Docker development:

```makefile
DOCKER_IMAGE ?= atoyr
DOCKER_TAG ?= latest

.PHONY: build up down clean

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

clean:
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true
```

### 3. `docker-compose.yml`

```yaml
services:
  atoyr:
    build:
      context: .
    image: atoyr:latest
    container_name: atoyr
    ports:
      - "80:80"
    env_file:
      - .env.production
    environment:
      PORT: 3000
      WORDS_PATH: data/words.json
      DATABASE_PATH: data/db.sqlite3
      ENV: production
    volumes:
      - ./packages/server/data:/app/server/data
```

Non-secret env vars are inlined in the compose file. Secrets (e.g., `GOOGLE_CLIENT_ID`, `ALLOWED_ADMIN_EMAILS`) go in a `.env.production` file at the project root, loaded via `env_file`. The SQLite database uses a bind mount to `./packages/server/data/` on the host, so the container reads/writes the same `db.sqlite3` file you use for local dev. The `words.json` and dashboard static files are shipped inside the image, not on a volume.

### 4. `.github/workflows/deploy.yml`

Single workflow that builds and pushes one image, then triggers a Coolify deploy webhook.

```yaml
name: Deploy

on:
  push:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - uses: actions/checkout@v4

      - name: Log in to GHCR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest

    steps:
      - name: Trigger Coolify deploy
        run: |
          curl -sS -X POST "${{ secrets.COOLIFY_DEPLOY_WEBHOOK }}" \
            -H "Content-Type: application/json" \
            -d '{}'
```

Key details:
- `packages: write` permission on the job — enables the built-in `GITHUB_TOKEN` to push to GHCR.
- Docker Buildx with GitHub Actions cache (`type=gha`) — avoids rebuilding layers from scratch on every push.
- Coolify deploy webhook is a secret (`COOLIFY_DEPLOY_WEBHOOK`) — a unique URL Coolify exposes per resource that triggers a re-deploy (pull + restart).

### 3. VM setup

- Run `docker login ghcr.io` with a classic PAT (scopes: `read:packages`, `repo`).
- Add a 2GB swap file (Coolify docs recommend this when builds share a server; cheap insurance for memory spikes from databases/backups/containers).
