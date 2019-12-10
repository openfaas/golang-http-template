package main

import (
	"context"
	"handler/function"
	"log"

	"github.com/cloudevents/sdk-go"
)

func main() {
	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithPort(8082),
		cloudevents.WithPath("/"))

	if err != nil {
		log.Fatal(err)
	}

	client, err := cloudevents.NewClient(t)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	log.Fatal(client.StartReceiver(ctx, function.Handle))
}
