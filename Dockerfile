# syntax=docker/dockerfile:1
# Build the manager binary
FROM golang:1.24 AS base
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY api/ api/
COPY internal/ internal/
COPY pkg/ pkg/
RUN CGO_ENABLED=0 GOOS=linux go build -o /dev/null ./internal/... ./pkg/... ./api/...

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
FROM base AS builder-controller
COPY cmd/controller cmd/controller
RUN CGO_ENABLED=0 GOOS=linux go build -o ./controller ./cmd/controller

FROM base AS builder-worker
COPY cmd/worker cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -o ./worker ./cmd/worker

FROM base AS builder-storage
COPY cmd/storage cmd/storage
RUN CGO_ENABLED=0 GOOS=linux go build -o ./storage ./cmd/storage

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot AS controller
WORKDIR /
COPY --from=builder-controller /workspace/controller .
USER 65532:65532

ENTRYPOINT ["/controller"]

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot AS storage
WORKDIR /
COPY --from=builder-storage /workspace/storage .
USER 65532:65532

ENTRYPOINT ["/storage"]

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot AS worker
WORKDIR /
COPY --from=builder-worker /workspace/worker .
USER 65532:65532

ENTRYPOINT ["/worker"]
