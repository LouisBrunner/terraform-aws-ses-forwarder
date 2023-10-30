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

	err = Handler(
		setupSession(), setupConfig(),
		events.SNSEvent{},
	)
	expectedError := "no record"
	assert.EqualError(t, err, expectedError)

	content, err := collect()
	if !assert.NoError(t, err) {
		return
	}
	assert.Regexp(t, "error: "+expectedError+"\n", string(content))
}

func TestHandler_Fails_Parsing(t *testing.T) {
	w, collect, err := iowrap.Writer()
	if !assert.NoError(t, err) {
		return
	}
	log.SetOutput(w)

	err = Handler(
		setupSession(), setupConfig(),
		events.SNSEvent{
			Records: []events.SNSEventRecord{
				{
					SNS: events.SNSEntity{
						MessageID: "123",
						Message:   "{,}",
					},
				},
			},
		},
	)
	expectedError := "123: invalid character ',' looking for beginning of object key string"
	assert.EqualError(t, err, expectedError)

	content, err := collect()
	if !assert.NoError(t, err) {
		return
	}
	assert.Regexp(t, "error: "+expectedError+"\n", string(content))
}

const eventValid = `{
  "content":"123",
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"},"recipients":["hello@moto.com"]},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`

func TestHandler_Fails_Mapping(t *testing.T) {
	w, collect, err := iowrap.Writer()
	if !assert.NoError(t, err) {
		return
	}
	log.SetOutput(w)

	err = Handler(
		setupSession(), setupConfig(),
		events.SNSEvent{
			Records: []events.SNSEventRecord{
				{
					SNS: events.SNSEntity{
						MessageID: "456",
						Message:   eventValid,
					},
				},
			},
		},
	)
	expectedError := "456: no destination"
	assert.EqualError(t, err, expectedError)

	content, err := collect()
	if !assert.NoError(t, err) {
		return
	}
	assert.Regexp(t, "error: "+expectedError+"\n", string(content))
}
