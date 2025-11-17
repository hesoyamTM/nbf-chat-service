FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    go build \
    -trimpath \
    -o /main \
    ./cmds/main.go

FROM scratch AS runtime

LABEL org.opencontainers.image.title="Neighbor Finder Chat Service" \
    org.opencontainers.image.description="Multi-platform Go application for Neighbor Finder Chat Service" \
    org.opencontainers.image.vendor="HesoyamTM" \
    org.opencontainers.image.version="$VERSION" \
    org.opencontainers.image.revision="$COMMIT"

COPY --from=builder /main /main

EXPOSE 50052

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/main", "health"]

ENTRYPOINT ["/main"]
