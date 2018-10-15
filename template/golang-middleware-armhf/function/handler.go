package function

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		input = body
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hello world, input was: %s", string(input))))
}
