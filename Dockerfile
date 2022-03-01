# Build the manager binary
FROM registry.access.redhat.com/ubi8/go-toolset:1.16.7-5 as builder
ARG BUILD_VERSION="unknown"

WORKDIR /opt/app-root

# TEMPORARY STEPS UNTIL ubi8 releases a go1.17 image
RUN wget -q https://go.dev/dl/go1.17.6.linux-amd64.tar.gz && tar -xzf go1.17.6.linux-amd64.tar.gz
ENV GOROOT /opt/app-root/go
ENV PATH $GOROOT/bin:$PATH
# END OF LINES TO REMOVE

# Copy the go manifests and source
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags "-X 'main.buildVersion=$BUILD_VERSION' -X 'main.buildDate=`date +%Y-%m-%d\ %H:%M`'" -mod vendor -a -o manager main.go

# Create final image from minimal + built binary
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.5-204
WORKDIR /
COPY --from=builder /opt/app-root/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
