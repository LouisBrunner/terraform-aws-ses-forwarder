package logic

import (
	"log"
	"testing"

	"github.com/LouisBrunner/go-iowrap"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

func setupSession() *session.Session {
	return session.Must(session.NewSession())
}

func TestHandler_Fails_NoBody(t *testing.T) {
	w, collect, err := iowrap.Writer()
	if !assert.NoError(t, err) {
		return
	}
	log.SetOutput(w)

	_, err = Handler(
		setupSession(), setupConfig(),
		events.APIGatewayProxyRequest{
			RequestContext: events.APIGatewayProxyRequestContext{
				RequestID: "123",
			},
		},
	)
	expectedError := "missing body"
	assert.EqualError(t, err, expectedError)

	content, err := collect()
	if !assert.NoError(t, err) {
		return
	}
	assert.Regexp(t, "123: error: "+expectedError+"\n", string(content))
}

func TestHandler_Fails_Parsing(t *testing.T) {
	w, collect, err := iowrap.Writer()
	if !assert.NoError(t, err) {
		return
	}
	log.SetOutput(w)

	_, err = Handler(
		setupSession(), setupConfig(),
		events.APIGatewayProxyRequest{
			RequestContext: events.APIGatewayProxyRequestContext{
				RequestID: "123",
			},
			Body: "{,}",
		},
	)
	expectedError := "invalid character ',' looking for beginning of object key string"
	assert.EqualError(t, err, expectedError)

	content, err := collect()
	if !assert.NoError(t, err) {
		return
	}
	assert.Regexp(t, "123: error: "+expectedError+"\n", string(content))
}

const eventValid = `{
  "content":"123",
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"}},
  "mail":{"headers":{"To":"hello@moto.com"}}
}`

func TestHandler_Fails_Mapping(t *testing.T) {
	w, collect, err := iowrap.Writer()
	if !assert.NoError(t, err) {
		return
	}
	log.SetOutput(w)

	_, err = Handler(
		setupSession(), setupConfig(),
		events.APIGatewayProxyRequest{
			RequestContext: events.APIGatewayProxyRequestContext{
				RequestID: "123",
			},
			Body: eventValid,
		},
	)
	expectedError := "no match found for `hello@moto.com`"
	assert.EqualError(t, err, expectedError)

	content, err := collect()
	if !assert.NoError(t, err) {
		return
	}
	assert.Regexp(t, "123: error: "+expectedError+"\n", string(content))
}
