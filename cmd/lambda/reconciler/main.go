package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handle(_ context.Context, event events.CloudWatchEvent) error {
	log.Printf("scheduled synthetic reconciliation event_id=%s", event.ID)
	return nil
}

func main() {
	lambda.Start(handle)
}
