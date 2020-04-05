# go-function-sdk
An SDK for building OpenFaaS functions in Go


## Installing

Use `go get` to retrieve the SDK to add it to your `GOPATH` workspace, or
project's Go module dependencies.

	go get github.com/openfaas-incubator/go-function-sdk

To update the SDK use `go get -u` to retrieve the latest version of the SDK.

	go get -u github.com/openfaas-incubator/go-function-sdk
	
	
## Features

### Handler definition
```go
type FunctionHandler interface {
	Handle(req Request) (Response, error)
}
```

`FunctionHandler` interface is used by [golang-http](https://github.com/openfaas-incubator/golang-http-template/tree/master/template/golang-http)
template to define a functions handler


### Secrets
For the time being please use the secrets function from `github.com/openfaas/openfaas-cloud/sdk`

See: https://github.com/openfaas/openfaas-cloud/blob/master/sdk/secrets.go

Usage:

```go
secret, err := sdk.ReadSecret("MY_SECRET")
if err != nil {
    return fmt.Errorf("error reading secret. %v", err)
}
```