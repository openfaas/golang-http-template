#!/bin/sh
# go.sh is a wrapper for the Go command that will
# automatically set the required module related flags
# when a vendor folder is detected.
#
# Currently, in Go 1.18, Go Workspaces are incompatible
# with vendored modules. As a result, we must disable
# Go modules and use GOPATH mode.

# We use this bash script to wrap Go commands because
# there is no clear way to set an env variable from the
# a script in a Dockerfile.
# It is possible to use  `go env -w` but, but env varaibles
# have precedence and if it is set as an arg/env variable,
# then it will be ignored by the `go env -w`.

# if the function/vendor folder exists
# then we set the env variables for
# GO111MODULE=off
if [ -d "/go/src/handler/function/vendor" ]; then
    echo "Setting vendor mode env variables"
    export GO111MODULE=off
fi

# if DEBUG env is 1, print the go env
if [ "${DEBUG:-0}" = "1" ]; then
    go env
fi

go "$@"
