# syntax=docker/dockerfile:1

# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.25.7-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

# Build the statically-linked binary. Migrations and the OpenAPI spec are
# embedded, so nothing else needs to ship in the final image.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/amana ./cmd/api

# ── Runtime stage ────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/amana /amana
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/amana"]
