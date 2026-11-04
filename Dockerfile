# syntax=docker/dockerfile:1
# BuildKit bind mounts keep compiler caches and node_modules out of build layers.
FROM node:24.22.0-alpine3.24 AS assets
RUN --mount=type=bind,target=/source,readonly \
    --mount=type=secret,id=proxy_ca \
    set -eu; \
    if [ -f /run/secrets/proxy_ca ]; then export NODE_EXTRA_CA_CERTS=/run/secrets/proxy_ca; fi; \
    cp -a /source /tmp/project; cd /tmp/project; \
    npm ci --no-audit --no-fund; npm run build; \
    mkdir -p /out; cp -r resources/public/build /out/build; cp -r resources/emails /out/emails; \
    rm -rf /tmp/project /root/.npm

FROM golang:1.27.3-alpine3.24 AS build
ENV CGO_ENABLED=0
RUN --mount=type=bind,target=/source,readonly \
    --mount=type=bind,from=assets,source=/out,target=/assets,readonly \
    --mount=type=secret,id=proxy_ca \
    set -eu; \
    if [ -f /run/secrets/proxy_ca ]; then export SSL_CERT_FILE=/run/secrets/proxy_ca; fi; \
    cp -a /source /tmp/project; cd /tmp/project; \
    cp -r /assets/build resources/public/build; cp -r /assets/emails/. resources/emails/; \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app; \
    rm -rf /tmp/project /go/pkg /root/.cache/go-build

FROM node:24.22.0-alpine3.24 AS runtime
WORKDIR /app
RUN --mount=type=bind,target=/source,readonly \
    --mount=type=bind,from=build,source=/out,target=/binary,readonly \
    --mount=type=bind,from=build,source=/etc/ssl/certs,target=/trust,readonly \
    --mount=type=secret,id=proxy_ca \
    set -eu; \
    if [ -f /run/secrets/proxy_ca ]; then export NODE_EXTRA_CA_CERTS=/run/secrets/proxy_ca; fi; \
    cp /source/package.json /source/package-lock.json ./; \
    npm ci --omit=dev --no-audit --no-fund; \
    mkdir -p scripts storage; \
    cp /source/scripts/email-renderer.tsx /source/scripts/render-email.tsx /source/scripts/isolated-mdx.ts scripts/; \
    cp /binary/app ./app; cp /trust/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt; \
    chown node:node storage; chmod 644 package.json package-lock.json scripts/*; rm -rf /root/.npm
USER node
ENV APP_ENV=production APP_DEBUG=false HOST=0.0.0.0 PORT=3000 LOG_FORMAT=json NODE_ENV=production MAIL_MAILER=resend QUEUE_CONNECTION=database QUEUE_CONCURRENCY=1 CACHE_STORE=database
EXPOSE 3000
ENTRYPOINT ["/app/app"]
