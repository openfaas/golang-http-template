FROM openfaas/of-watchdog:0.8.0 AS watchdog

#
# ---
#

FROM golang:1.15.0-alpine3.12 AS build

ARG UID=1000
ARG GID=1000

# ignore non-pinned packages
# hadolint ignore=DL3018
RUN apk --no-cache add git tzdata ca-certificates \
    && addgroup -g "$GID" -S app && adduser -u "$UID" -S -D -H -g app app

ENV CGO_ENABLED=0

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

RUN mkdir -p /go/src/handler
WORKDIR /go/src/handler
COPY . .

# Run a gofmt and exclude all vendored code.
# hadolint ignore=SC2046
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./function/vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

ARG GO111MODULE="off"
ARG GOPROXY=""

RUN go build --ldflags '-s -w -extldflags "-static"' -a -installsuffix cgo -o handler .
RUN go test handler/function/... -cover

#
# ---
#

FROM scratch

# run out of /home/app by default
WORKDIR /home/app

# add-in our timezone data file
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

# add-in our unprivileged user
COPY --from=build /etc/passwd /etc/group /etc/shadow /etc/

# add-in our ca certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# from now on, run as the unprivileged user
USER app

# copy-in our stuff
COPY --from=build /go/src/handler/handler    .
COPY --from=build /usr/bin/fwatchdog         .
COPY --from=build /go/src/handler/function/  .

# set our variables
ENV fprocess="./handler"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:8082"

# run our watchdog in the end
CMD ["./fwatchdog"]
