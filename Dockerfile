# syntax = docker/dockerfile:1.2
ARG GO_VERSION=1.16

# get modules, if they don't change the cache can be used for faster builds
FROM golang:${GO_VERSION}-alpine AS base
ENV GO111MODULE=on
ENV CGO_ENABLED=0
WORKDIR /src
COPY go.* .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM golang:${GO_VERSION}-alpine AS dlv
ENV GO111MODULE=on
ENV GOPATH=/go
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# build the dev instance containing all source files for file sync + dlv for remote debugging
# from base to cache the modules
FROM base as dev
COPY --from=cosmtrek/air /go/bin/air /usr/local/bin
COPY --from=dlv /go/bin/dlv /usr/local/bin
COPY . .
ENTRYPOINT ["air", "-c", ".air.toml"]

# build the application itself
FROM base AS build
# temp mount all files instead of loading into image with COPY
# temp mount module cache
# temp mount go build cache
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/main ./cmd/noa/main.go

# Import the binary from build stage
FROM gcr.io/distroless/static:nonroot as prd
COPY --from=build /app/main /
USER nonroot:nonroot
ENTRYPOINT ["/main"]
