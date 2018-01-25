package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"handler/function"
)

func makeRequestHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var input []byte

		if r.Body != nil {

			defer r.Body.Close()

			bodyBytes, bodyErr := ioutil.ReadAll(r.Body)

			if bodyErr != nil {
				log.Printf("Error reading body from request.")
			}

			input = bodyBytes
		}

		result := function.Handle(input)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
	}
}

func main() {
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8081),
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   3 * time.Second,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	http.HandleFunc("/", makeRequestHandler())
	log.Fatal(s.ListenAndServe())
}
