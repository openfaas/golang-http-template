package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"handler/function"
	//"github.com/openfaas-incubator/golang-http-template/template/golang-middleware/function"
)

func main() {
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8081),
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   3 * time.Second,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	http.HandleFunc("/", function.Handle)
	log.Fatal(s.ListenAndServe())
}
