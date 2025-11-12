package function

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"net/http"
	"time"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	var input []byte

	log := slog.New(slog.NewTextHandler(os.Stdout, opts))
	log = log.With("X-Call-Id", r.Header.Get("X-Call-Id"))

	log.Info("received request")

	if r.Body != nil {
		defer r.Body.Close()

		body, _ := io.ReadAll(r.Body)

		input = body
	}

	log.Info("Sleeping for 1 milliseconds")

	time.Sleep(time.Millisecond * 1)

	log.Info("Sleep done")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Body: %s", string(input))))
}
