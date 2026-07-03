# ---- Client build ----
FROM node:24-alpine AS client-build

WORKDIR /app

COPY package.json yarn.lock tsconfig.json .yarnrc.yml ./
COPY packages/client/package.json packages/client/package.json

RUN corepack enable
RUN yarn workspaces focus client

COPY packages/client/ packages/client/

RUN yarn workspace client build

# ---- Server build ----
FROM golang:1.25-alpine AS server-build

WORKDIR /app

COPY packages/server/go.mod packages/server/go.sum ./
RUN go mod download

COPY packages/server/ .

RUN apk add --no-cache build-base
ARG GIT_HASH=unknown
RUN CGO_ENABLED=1 go build -ldflags "-X main.GitHash=${GIT_HASH}" -o /app/server ./cmd/service

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
