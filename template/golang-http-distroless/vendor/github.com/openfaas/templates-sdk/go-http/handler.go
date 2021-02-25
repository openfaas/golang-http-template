// Copyright (c) Alex Ellis 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package handler

import (
	"context"
	"net/http"
)

// Response of function call
type Response struct {

	// Body the body will be written back
	Body []byte

	// StatusCode needs to be populated with value such as http.StatusOK
	StatusCode int

	// Header is optional and contains any additional headers the function response should set
	Header http.Header
}

// Request of function call
type Request struct {
	Body        []byte
	Header      http.Header
	QueryString string
	Method      string
	Host        string
	ctx         context.Context
}

// Context is set for optional cancellation inflight requests.
func (r *Request) Context() context.Context {
	return r.ctx
}

// WithContext overides the context for the Request struct
func (r *Request) WithContext(ctx context.Context) {
	// AE: Not keen on panic mid-flow in user-code, however stdlib also appears to do
	// this. https://golang.org/src/net/http/request.go
	// This is not setting a precedent for broader use of "panic" to handle errors.
	if ctx == nil {
		panic("nil context")
	}
	r.ctx = ctx
}

// FunctionHandler used for a serverless Go method invocation
type FunctionHandler interface {
	Handle(req Request) (Response, error)
}

func init() {

}
