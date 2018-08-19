package main

import (
	"github.com/LouisBrunner/aws-ses-forwarder/logic"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	sess := session.Must(session.NewSession())
	config, err := logic.LoadConfig("./ef.json")
	if err != nil {
		panic("could not load config: " + err.Error())
	}
	lambda.Start(func(event events.SNSEvent) error {
		return logic.Handler(sess, config, event)
	})
}
