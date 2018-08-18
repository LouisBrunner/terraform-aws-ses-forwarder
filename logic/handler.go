package logic

import (
	"errors"
	"log"

	"github.com/LouisBrunner/aws-ses-forwarder/mailer"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Handler is the Lambda function handler
// It uses Amazon API Gateway request/responses provided by the aws-lambda-go/events package,
func Handler(session client.ConfigProvider, conf *Config, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, ferr error) {
	log.Printf("%s: start processing\n", request.RequestContext.RequestID)
	defer func() {
		if ferr != nil {
			log.Printf("%s: error: %+v\n", request.RequestContext.RequestID, ferr)
		}
		log.Printf("%s: done processing\n", request.RequestContext.RequestID)
	}()

	if len(request.Body) < 1 {
		return response, errors.New("missing body")
	}

	defer log.Printf("%s: parse body\n", request.RequestContext.RequestID)
	email, err := mailer.ParseEvent([]byte(request.Body))
	if err != nil {
		return response, err
	}

	defer log.Printf("%s: map destination (to: %s)\n", request.RequestContext.RequestID, email.To)
	to, err := conf.Map(email.To)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	defer log.Printf("%s: forward to %s\n", request.RequestContext.RequestID, to)
	err = email.Forward(session, to)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       "Forwarded",
		StatusCode: 200,
	}, nil
}
