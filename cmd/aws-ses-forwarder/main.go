package main

import (
	"os"

	"github.com/LouisBrunner/aws-ses-forwarder/pkg/logic"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	sess := session.Must(session.NewSession())
	rawConfig := os.Getenv("CONFIG")
	if rawConfig == "" {
		panic("CONFIG environment variable is required")
	}
	config, err := logic.LoadConfig(rawConfig)
	if err != nil {
		panic("could not load config: " + err.Error())
	}
	lambda.Start(func(event events.SNSEvent) error {
		return logic.Handler(sess, config, event)
	})
}
