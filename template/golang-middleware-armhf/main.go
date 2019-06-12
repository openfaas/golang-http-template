package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"handler/function"
	//"github.com/openfaas-incubator/golang-http-template/template/golang-middleware/function"
)

func main() {
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8082),
		ReadTimeout:    mustParseDuration(os.Getenv("read_timeout")),
		WriteTimeout:   mustParseDuration(os.Getenv("write_timeout")),
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	http.HandleFunc("/", function.Handle)
	log.Fatal(s.ListenAndServe())
}

func mustParseDuration(v string) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		panic(err)
	}
	return d
}
