FROM --platform=${TARGETPLATFORM:-linux/amd64} ghcr.io/openfaas/of-watchdog:0.8.4 as watchdog
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.15-alpine3.13 as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk --no-cache add git

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV CGO_ENABLED=0

RUN mkdir -p /go/src/handler
WORKDIR /go/src/handler
COPY . .

ARG GO111MODULE="off"
ARG GOPROXY=""
ARG GOFLAGS=""

# Add user overrides to the root go.mod, which is the only place "replace" can be used
RUN cat function/GO_REPLACE.txt >> ./go.mod || exit 0
RUN if [ -d ./function/vendor ]; then \
    echo "moving handler vendor" && \
    mv -f ./function/vendor .;\
    else \
    echo "vendor not found "; \
    fi

RUN if [ "${GO111MODULE}" == 'on' ]; then \
        # copy the user's go.mod
        mv -f ./function/go.mod . && \
        mv -f ./function/go.sum . && \
        # clean up the go.mod,
        # remove any local require for the handler/function, this can exist _if_ the
        # developer has subpackages in their handler. It is ok to just remove it
        # because we will replace it later
        grep -v "handler/function" go.mod > gomod2; mv gomod2 go.mod && \
        # now update the go.mod
        # first, replace handler/function to point at the local code
        # second, we need to rename the module to handler to match our main
        go mod edit \
            -replace=handler/function=./function \
            -module handler && \
        cat go.mod && \
        if [ -d ./vendor ]; then \
            # when vendored, we need to do similar edits to the vendor/modules.txt
            # first we need to replace any possible copy of the handler code
            rm -rf vendor/handler && \
            # in modules.txt, we remove existing references to the handler/function
            # these are replaced later with the new structure
            grep -v "handler/function" ./vendor/modules.txt> modulestext; mv modulestext ./vendor/modules.txt && \
            # add the mising replace to the vendor/modules.txt
            echo "## explicit" >> ./vendor/modules.txt && \
            echo "# handler/function => ./function" >> ./vendor/modules.txt && \
            cat ./vendor/modules.txt; \
        else \
            echo "skip adding replace to ' ./vendor/modules.txt'"; \
        fi \
    else \
        echo "skip go.mod handling"; \
    fi

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./function/vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }


WORKDIR /go/src/handler/function

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go test ./... -cover

WORKDIR /go/src/handler
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build --ldflags "-s -w" -a -installsuffix cgo -o handler .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.13
# Add non root user and certs
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app
# Split instructions so that buildkit can run & cache
# the previous command ahead of time.
RUN mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app

COPY --from=build --chown=app /go/src/handler/handler    .
COPY --from=build --chown=app /usr/bin/fwatchdog         .
COPY --from=build --chown=app /go/src/handler/function/  .

USER app

ENV fprocess="./handler"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:8082"
ENV prefix_logs="false"

CMD ["./fwatchdog"]
