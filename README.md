# OpenFaaS Golang HTTP templates

This repository contains two Golang templates for OpenFaaS which give additional control over the HTTP request and response. They will both handle higher throughput than the classic watchdog due to the process being kept warm.

Our recommended template for Go developers is golang-middleware.

You'll find a chapter dedicated to writing functions with Go in [Everyday Golang by Alex Ellis](https://store.openfaas.com/l/everyday-golang)

Using the templates:

```bash
faas-cli template store pull golang-http
faas-cli template store pull golang-middleware
```

Or:

```bash
$ faas template pull https://github.com/openfaas/golang-http-template
$ faas new --list

Languages available as templates:
- golang-http
- golang-middleware
```

The two templates are very similar:

* `golang-middleware` implements a `http.HandleFunc` from Go's stdlib.
* `golang-http` uses a structured request/response object

## Dependencies

You can manage dependencies in one of the following ways:

- To use Go modules without vendoring, the default already is set `GO111MODULE=on` but you also can make that explicit by adding `--build-arg GO111MODULE=on` to `faas-cli up`, you can also use `--build-arg GOPROXY=https://` if you want to use your own mirror for the modules
- You can also Go modules with vendoring, run `go mod vendor` in your function folder and add `--build-arg GO111MODULE=off --build-arg GOFLAGS='-mod=vendor'` to `faas-cli up`
- If you have a private module dependency, we recommend using the vendoring technique from above.

### SSH authentication for private Git repositories and modules

If you do not wish to, or cannot use vendoring for some reason, then we provide an alternative set of templates for OpenFaaS Pro customers:

* [OpenFaaS Pro templates for Go](https://github.com/openfaas/pro-templates)

## 1.0 golang-middleware (recommended template)

This is one of the fastest templates available for Go available. Its signature is a [http.HandlerFunc](https://golang.org/pkg/net/http/#HandlerFunc), instead of a traditional request and response that you may expect from a function.

The user has complete control over the HTTP request and response.

### Get the template

```
$ faas template store pull golang-middleware

# Or

$ faas template pull https://github.com/openfaas/golang-http-template
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

### Adding static files to your image

If a folder named `static` is found in the root of your function's source code, **it will be copied** into the final image published for your function.

To read this back at runtime, you can do the following:

```go
package function

import (
    "net/http"
    "os"
)

func Handle(w http.ResponseWriter, r *http.Request) {

    data, err := os.ReadFile("./static/file.txt")

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    w.Write(data)
}
```

## 2.0 golang-http

This template provides additional context and control over the HTTP response from your function.

### Status of the template

Like the `golang-middleware` template, this template is highly performant and suitable for production.

### Get the template

```sh
$ faas template store pull golang-http

# Or
$ faas template pull https://github.com/openfaas/golang-http-template
$ faas new --lang golang-http <fn-name>
```

### Example usage

Example writing a successful message:

```go
package function

import (
	"fmt"
	"net/http"

		handler "github.com/openfaas/templates-sdk/go-http"
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

		handler "github.com/openfaas/templates-sdk/go-http"
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

		handler "github.com/openfaas/templates-sdk/go-http"
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

		handler "github.com/openfaas/templates-sdk/go-http"
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

	handler "github.com/openfaas/templates-sdk/go-http"
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

