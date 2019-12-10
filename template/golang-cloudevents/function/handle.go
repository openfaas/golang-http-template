package function

import (
	"log"

	"github.com/cloudevents/sdk-go"
)

func Handle(event cloudevents.Event) {
	log.Printf(`event received: (type: %s, ID: %s, source: %s)`, event.Type(), event.ID(), event.Source())
}
