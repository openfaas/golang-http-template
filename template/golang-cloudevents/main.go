package main

import (
	"context"
	"handler/function"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	t, err := cloudevents.NewHTTP(
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
