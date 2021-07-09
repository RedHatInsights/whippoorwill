# Build the manager binary
FROM registry.access.redhat.com/ubi8/go-toolset:1.15.7 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
USER 0
RUN go mod download

COPY Makefile Makefile

RUN make controller-gen kustomize

COPY hack/boilerplate.go.txt hack/boilerplate.go.txt

# Copy the go source
COPY main.go main.go
COPY config/ config/
COPY apis/ apis/
COPY controllers/ controllers/

RUN make manifests generate fmt vet release

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/manifest.yaml .
USER 65532:65532

ENTRYPOINT ["/manager"]
