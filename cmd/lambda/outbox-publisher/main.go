package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handle(_ context.Context, event events.DynamoDBEvent) error {
	for _, record := range event.Records {
		log.Printf("observed outbox stream event id=%s name=%s", record.EventID, record.EventName)
	}
	return nil
}

func main() {
	lambda.Start(handle)
}
