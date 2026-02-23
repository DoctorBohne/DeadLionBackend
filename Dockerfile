# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.24.3

########################
# 1) Build stage
########################
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /src

# Für reproduzierbare, kleine Binaries (wenn dein Projekt kein CGO braucht)
ENV CGO_ENABLED=0

# Buildx/Multiplattform: nutzt automatisch TARGETOS/TARGETARCH, wenn gesetzt
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS:-linux}
ENV GOARCH=${TARGETARCH:-amd64}

# Abhängigkeiten zuerst kopieren (bessere Layer-Caches)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Restlicher Source
COPY . .

# Pfad zum main-Package (bei Bedarf beim Build überschreiben)
ARG APP_PATH=./cmd/server

# Binary bauen
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/app ${APP_PATH}

########################
# 2) Runtime stage
########################
# Sehr kleines, sicheres Runtime-Image (ohne Shell)
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /
COPY --from=builder /out/app /app

# Optional: wenn dein Service einen Port hat
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app"]