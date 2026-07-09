package main

import (
	"context"
	"os"

	"github.com/LouisBrunner/aws-ses-forwarder/pkg/logic"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("could not load AWS config: " + err.Error())
	}
	rawConfig := os.Getenv("CONFIG")
	if rawConfig == "" {
		panic("CONFIG environment variable is required")
	}
	conf, err := logic.LoadConfig(rawConfig)
	if err != nil {
		panic("could not load config: " + err.Error())
	}
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		return logic.Handler(ctx, cfg, conf, event)
	})
}
