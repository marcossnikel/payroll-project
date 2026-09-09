package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handle(_ context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		log.Printf("received synthetic payment command message_id=%s", record.MessageId)
	}
	return nil
}

func main() {
	lambda.Start(handle)
}
