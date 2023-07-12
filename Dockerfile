# We do not use --platform feature to auto fill this ARG because of incompatibility between podman and docker
ARG TARGETPLATFORM=linux/amd64
ARG BUILDPLATFORM=linux/amd64
# Build the manager binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.20 as builder
ARG BUILD_VERSION="unknown"

ARG TARGETPLATFORM
ARG TARGETARCH=amd64
WORKDIR /opt/app-root

# Copy the go manifests and source
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags "-X 'main.buildVersion=$BUILD_VERSION' -X 'main.buildDate=`date +%Y-%m-%d\ %H:%M`'" -mod vendor -a -o manager main.go

# Create final image from minimal + built binary
FROM --platform=$TARGETPLATFORM registry.access.redhat.com/ubi9/ubi-minimal:9.2
WORKDIR /
COPY --from=builder /opt/app-root/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
