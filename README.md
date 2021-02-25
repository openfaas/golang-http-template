OpenFaaS Golang HTTP templates - using distroless as base
=============================================

This repository contains two Golang templates for OpenFaaS which give additional control over the HTTP request and response. They will both handle higher throughput than the classic watchdog due to the process being kept warm.

## Diff from upstream - https://github.com/openfaas/golang-http-template


Using the templates:

```bash
faas-cli template pull https://github.com/lexxxel/golang-distroless-template
```

Or:

```bash
$ faas new --list

Languages available as templates:
- golang-http-distroless
- golang-middleware-distroless
```

The two templates are equivalent with `golang-http-distroless` using a structured request/response object and the alternative implementing a Golang `http.HandleFunc` from the Golang stdlib. `golang-http-distroless` is more "conventional" for a Golang serverless template but this is a question of style/taste.

## Dependencies

You can manage dependencies in one of the following ways:

* To use Go modules without vendoring, add `--build-arg GO111MODULE=on` to `faas-cli up`, you can also use `--build-arg GOPROXY=https://` if you want to use your own mirror for the modules
* For traditional vendoring with `dep` give no argument, or add `--build-arg GO111MODULE=off` to `faas-cli up`

## 1.0 golang-http-distroless

This template provides additional context and control over the HTTP response from your function.

### Status of the template

This template is the most performant and recent Golang template for OpenFaaS which also provides a function-style request and response for the user.

### Get the template

```sh
$ faas template pull https://github.com/lexxxel/golang-distroless-template
$ faas new --lang golang-http-distroless <fn-name>
```

### Example usage

Example writing a successful message:

```go
package function

import (
	"fmt"
	"net/http"

	"github.com/openfaas-incubator/go-function-sdk"
)

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {
	var err error

	message := fmt.Sprintf("Hello world, input was: %s", string(req.Body))

	return handler.Response{
		Body:       []byte(message),
    }, err
}
```

Example writing a custom status code

```go
package function

import (
	"fmt"
	"net/http"

	"github.com/openfaas-incubator/go-function-sdk"
)

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {
	var err error

	return handler.Response{
		Body:       []byte("Your workload was accepted"),
		StatusCode: http.StatusAccepted,
	}, err
}
```

Example writing an error / failure.

```go
package function

import (
	"fmt"
	"net/http"

	"github.com/openfaas-incubator/go-function-sdk"
)

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {
	var err error

	return handler.Response{
        Body: []byte("the input was invalid")
	}, fmt.Errorf("invalid input")
}
```

The error will be logged to `stderr` and the `body` will be written to the client along with a HTTP 500 status code.

Example reading a header.

```go
package function

import (
	"log"

	"github.com/openfaas-incubator/go-function-sdk"
)

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {
	var err error

	log.Println(req.Header) // Check function logs for the request headers

	return handler.Response{
		Body: []byte("This is the response"),
		Header: map[string][]string{
			"X-Served-By": []string{"My Awesome Function"},
		},
	}, err
}
```

Example responding to an aborted request.

The `Request` object provides access to the request context. This allows you to check if the request has been cancelled by using the context's done channel `req.Context().Done()` or the context's error `req.Context().Err()`

```go
package function

import (
	"fmt"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
)

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {
	var err error

	for i := 0; i < 10000; i++ {
		if req.Context().Err() != nil  {
			return handler.Response{}, fmt.Errorf("request cancelled")
		}
		fmt.Printf("count %d\n", i)
	}

	message := fmt.Sprintf("Hello world, input was: %s", string(req.Body))
	return handler.Response{
		Body:       []byte(message),
		StatusCode: http.StatusOK,
	}, err
}
```

This context can also be passed to other methods so that they can respond to the cancellation as well, for example [`db.ExecContext(req.Context())`](https://golang.org/pkg/database/sql/#DB.ExecContext)


## 2.0 golang-middleware-distroless

This template uses the [http.HandlerFunc](https://golang.org/pkg/net/http/#HandlerFunc) as entry point.

### Status of the template

Like the golang-http-distroless template, this is one of the fastest templates available, but takes a more service-orientated approach to its signature. Instead of looking like a traditional function, the user has complete control over the HTTP request and response.

### Get the template

```
$ faas template store pull golang-middleware-distroless

# Or

$ faas template pull https://github.com/lexxxel/golang-distroless-template
$ faas new --lang golang-middleware-distroless <fn-name>
```

### Example usage

Example writing a JSON response:

```go
package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		// read request payload
		reqBody, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		input = reqBody
		}
	}

	// log to stdout
	fmt.Printf("request body: %s", string(input))

	response := struct {
		Payload     string              `json:"payload"`
		Headers     map[string][]string `json:"headers"`
		Environment []string            `json:"environment"`
	}{
		Payload:     string(input),
		Headers:     r.Header,
		Environment: os.Environ(),
	}

	resBody, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    // write result
	w.WriteHeader(http.StatusOK)
	w.Write(resBody)
}
```

Example persistent database connection pool between function calls:

```go
package function

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	_ "github.com/go-sql-driver/mysql"
)

// db pool shared between function calls
var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		panic(err.Error())
	}

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
}

func Handle(w http.ResponseWriter, r *http.Request) {
	var query string
	ctx := r.Context()

	if r.Body != nil {
		defer r.Body.Close()

		// read request payload
		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		query = string(body)
	}

	// log to stdout
	fmt.Printf("Executing query: %s", query)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		if e := ctx.Err(); e != nil {
			http.Error(w, e, http.StatusBadRequest)
			return
		}
		var id int
		if err := rows.Scan(&id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ids = append(ids, string(id))
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := fmt.Sprintf("ids %s", strings.Join(ids, ", "))

	// write result
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}
```

Example retrieving request query strings

```go
package function
import (
	"fmt"
	"net/http"
)
func Handle(w http.ResponseWriter, r *http.Request) {
	// Parses RawQuery and returns the corresponding
	// values as a map[string][]string
	// for more info https://golang.org/pkg/net/url/#URL.Query
	query := r.URL.Query()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("id: %s", query.Get("id"))))
}
```

#### Advanced usage - Go sub-modules via `GO_REPLACE.txt`

For this example you will need to be using Go 1.13 or newer and Go modules, enable this via `faas-cli build --build-arg GO111MODULE=on`.

Imagine you have a package which you want to store outside of the `handler.go` file, it's another middleware which can perform an echo of the user's input.

```Golang
package handlers

import (
	"io/ioutil"
	"net/http"
)

func Echo(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
		b, _ := ioutil.ReadAll(r.Body)
		w.Write(b)
	}
}
```

To include a relative module such as this new `handlers` package, you should create a `GO_REPLACE.txt` file as follows:

```
replace github.com/alexellis/vault/purchase/handlers => ./function/handlers
```

How did we get to that? Let's say your GOPATH for your GitHub repo is: `github.com/alexellis/vault/` and your OpenFaaS function is `purchase` generated by `faas-cli new purchase --lang golang-middleware-distroless`.

Your relative GOPATH is: `github.com/alexellis/vault/purchase`, so add a redirect as per below to redirect the "handlers" package to where it exists in the build container.


Now if you want to reference the handlers package from within your `handler.go` write the following:

```golang
package function

import (
	"net/http"

	"github.com/alexellis/vault/purchase/handlers"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	handlers.Echo(w, r)
}
```
