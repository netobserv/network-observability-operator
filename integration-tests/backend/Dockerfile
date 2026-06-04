ARG TARGETARCH

FROM docker.io/library/golang:1.25 AS builder

ARG TARGETARCH=amd64

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Copy the source code (changes frequently - separate layer)
COPY *.go ./
COPY testdata/ testdata/

# Build ginkgo test binary (only recompiles test code, deps already cached)
RUN GOGC=50 GOMAXPROCS=4 GOOS=linux GOARCH=$TARGETARCH \
    ginkgo build --ldflags="-s -w" -o e2e-tests.test

FROM --platform=linux/$TARGETARCH registry.access.redhat.com/ubi9/ubi-minimal:9.8-1777460003

WORKDIR /tests

# Copy the pre-compiled test binary, testdata and ginkgo binary
COPY --from=builder /workspace/e2e-tests.test .
COPY --from=builder /workspace/testdata/ testdata/
COPY --from=builder /go/bin/ginkgo /usr/local/bin/ginkgo

USER 65532:65532

# Default entrypoint uses ginkgo CLI pointing to pre-compiled binary
# This enables clean ginkgo syntax without -ginkgo. prefix
ENTRYPOINT ["ginkgo", "run", "./e2e-tests.test"]
