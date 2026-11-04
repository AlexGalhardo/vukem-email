# syntax=docker/dockerfile:1
FROM node:24.22.0-alpine3.24 AS assets
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY . .
RUN npm run build

FROM golang:1.27.3-alpine3.24 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=assets /app/resources/public/build ./resources/public/build
COPY --from=assets /app/resources/emails ./resources/emails
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app

FROM node:24.22.0-alpine3.24 AS runtime
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev --no-audit --no-fund && mkdir -p scripts storage && chown node:node storage
COPY scripts/email-renderer.tsx scripts/render-email.tsx scripts/isolated-mdx.ts ./scripts/
COPY --from=build /out/app ./app
USER node
ENV APP_ENV=production APP_DEBUG=false HOST=0.0.0.0 PORT=3000 LOG_FORMAT=json NODE_ENV=production MAIL_MAILER=resend QUEUE_CONNECTION=database QUEUE_CONCURRENCY=1 CACHE_STORE=database
EXPOSE 3000
ENTRYPOINT ["/app/app"]
