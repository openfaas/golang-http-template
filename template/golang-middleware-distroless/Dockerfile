FROM --platform=${TARGETPLATFORM:-linux/amd64} ghcr.io/openfaas/of-watchdog:0.8.4 as watchdog
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.16-alpine3.13 as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog && \
    apk --no-cache add git

ENV CGO_ENABLED=0
ARG GOPROXY=""
ARG GOFLAGS=""

RUN mkdir -p /go/src/handler
WORKDIR /go/src/handler
COPY . .

# Add user overrides to the root go.mod, which is the only place "replace" can be used
RUN cat function/GO_REPLACE.txt >> ./go.mod || exit 0 && \
    go get handler/function && \
    # Run a gofmt and exclude all vendored code. \
    test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./function/vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; } && \
    CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build --ldflags "-s -w" -a -installsuffix cgo -o handler . && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go test handler/function/... -cover

FROM --platform=${TARGETPLATFORM:-linux/amd64} gcr.io/distroless/base

COPY --from=build --chown=nonroot:nonroot /go/src/handler/handler    /
COPY --from=build --chown=nonroot:nonroot /usr/bin/fwatchdog         /

USER nonroot

ENV fprocess="/handler"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:8082"
ENV prefix_logs="false"

CMD ["/fwatchdog"]
