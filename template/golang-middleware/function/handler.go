package function

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	// read request payload
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// log to stdout
	fmt.Printf("request payload: %s", string(body))

	// write result
	message := fmt.Sprintf("Hello world, input was: %s", string(body))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}
