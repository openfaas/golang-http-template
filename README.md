# OpenFaaS Golang HTTP templates

This repository contains two Golang templates for OpenFaaS which give additional control over the HTTP request and response. They will both handle higher throughput than the classic watchdog due to the process being kept warm.

Using the templates:

```bash
faas-cli template store pull golang-http
faas-cli template store pull golang-middleware
```

Or:

```bash
$ faas template pull https://github.com/openfaas-incubator/golang-http-template
$ faas new --list

Languages available as templates:
- golang-http
- golang-middleware
```

The two templates are equivalent with `golang-http` using a structured request/response object and the alternative implementing a Golang `http.HandleFunc` from the Golang stdlib. `golang-http` is more "conventional" for a Golang serverless template but this is a question of style/taste.

## Dependencies

You can manage dependencies in one of the following ways:

- To use Go modules without vendoring, the default already is set `GO111MODULE=on` but you also can make that explicit by adding `--build-arg GO111MODULE=on` to `faas-cli up`, you can also use `--build-arg GOPROXY=https://` if you want to use your own mirror for the modules
- You can also Go modules with vendoring, run `go mod vendor` in your function folder and add `--build-arg GO111MODULE=off --build-arg GOFLAGS='-mod=vendor'` to `faas-cli up`
- If you have a private module dependency, we recommend using the vendoring technique from above.

## 1.0 golang-http

This template provides additional context and control over the HTTP response from your function.

### Status of the template

This template is the most performant and recent Golang template for OpenFaaS which also provides a function-style request and response for the user.

### Get the template

```sh
$ faas template store pull golang-http

# Or
$ faas template pull https://github.com/openfaas-incubator/golang-http-template
$ faas new --lang golang-http <fn-name>
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

## 2.0 golang-middleware

This template uses the [http.HandlerFunc](https://golang.org/pkg/net/http/#HandlerFunc) as entry point.

### Status of the template

Like the golang-http template, this is one of the fastest templates available, but takes a more service-orientated approach to its signature. Instead of looking like a traditional function, the user has complete control over the HTTP request and response.

### Get the template

```
$ faas template store pull golang-middleware

# Or

$ faas template pull https://github.com/openfaas-incubator/golang-http-template
$ faas new --lang golang-middleware <fn-name>
```

### Example usage

Example writing a JSON response:

```go
package function

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		// read request payload
		reqBody, err := io.ReadAll(r.Body)

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
	"io"
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
		body, err := io.ReadAll(r.Body)

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

#### Advanced usage

##### Sub-packages

It is often natural to organize your code into sub-packages, for example you may have a function with the following structure

```
./echo
├── go.mod
├── go.sum
├── handler.go
└── pkg
    └── version
        └── version.go
```

Now if you want to reference the`version` sub-package, import it as

```go
import "handler/function/pkg/version"
```

This works like any local Go project.

##### Go sub-modules

Sub-modules (meaning sub-folders with a `go.mod`) are not supported.
