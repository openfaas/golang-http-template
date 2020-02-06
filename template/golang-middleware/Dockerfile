FROM openfaas/of-watchdog:0.7.6 as watchdog
FROM golang:1.13-alpine3.11 as build

RUN apk --no-cache add git

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV CGO_ENABLED=0

RUN mkdir -p /go/src/handler
WORKDIR /go/src/handler
COPY . .

# Add user overrides to the root go.mod, which is the only place "replace" can be used
RUN cat function/GO_REPLACE.txt >> ./go.mod || exit 0

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./function/vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

ARG GO111MODULE="off"
ARG GOPROXY=""
ARG GOFLAGS=""

WORKDIR /go/src/handler/function

RUN go test ./... -cover

WORKDIR /go/src/handler
RUN go build --ldflags "-s -w" -a -installsuffix cgo -o handler .

FROM alpine:3.11
# Add non root user and certs
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app

COPY --from=build --chown=app /go/src/handler/handler    .
COPY --from=build --chown=app /usr/bin/fwatchdog         .
COPY --from=build --chown=app /go/src/handler/function/  .

USER app

ENV fprocess="./handler"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:8082"

CMD ["./fwatchdog"]
