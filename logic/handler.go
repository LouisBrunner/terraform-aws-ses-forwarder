package logic

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/LouisBrunner/aws-ses-forwarder/mailer"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Handler is the Lambda function handler
// It uses Amazon API Gateway request/responses provided by the aws-lambda-go/events package,
func Handler(session client.ConfigProvider, conf *Config, event events.SNSEvent) (ferr error) {
	log.Printf("start processing\n")
	defer func() {
		if ferr != nil {
			log.Printf("error: %+v\n", ferr)
		}
		log.Printf("done processing\n")
	}()

	if len(event.Records) < 1 {
		return errors.New("no record")
	}

	errorsList := []string{}

	for _, record := range event.Records {
		err := handleRecord(session, conf, &record)
		if err != nil {
			errorsList = append(errorsList, fmt.Sprintf("%s: %v", record.SNS.MessageID, err))
		}
	}

	if len(errorsList) > 0 {
		return errors.New(strings.Join(errorsList, ", "))
	}
	return nil
}

func handleRecord(session client.ConfigProvider, conf *Config, record *events.SNSEventRecord) error {
	log.Printf("new record: %+v\n", record)

	log.Printf("%s: parse body\n", record.SNS.MessageID)
	email, err := mailer.ParseEvent([]byte(record.SNS.Message))
	if err != nil {
		return err
	}

	log.Printf("%s: map destination (to: %s)\n", record.SNS.MessageID, email.To)
	to, err := conf.Map(email.To)
	if err != nil {
		return err
	}

	log.Printf("%s: forward to %s\n", record.SNS.MessageID, to)
	return email.Forward(session, to)
}
