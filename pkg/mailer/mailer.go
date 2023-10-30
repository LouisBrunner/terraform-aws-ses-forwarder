package mailer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ses"
)

const passVerdict = "PASS"

type raw struct {
	events.SimpleEmailService
	Content string `json:"content"`
}

// Event contains a SES event
type Event struct {
	To    []string
	email []byte
}

// ParseEvent will try to transform the argument in a SES event
func ParseEvent(rawJSON []byte) (*Event, error) {
	var rawEvent raw
	err := json.Unmarshal(rawJSON, &rawEvent)
	if err != nil {
		return nil, err
	}

	if len(rawEvent.Content) < 1 {
		return nil, errors.New("missing `content` in SES event")
	}
	if rawEvent.Receipt.SpamVerdict.Status != passVerdict || rawEvent.Receipt.VirusVerdict.Status != passVerdict {
		return nil, errors.New("don't forward spam/virus")
	}
	if len(rawEvent.Receipt.Recipients) < 1 {
		return nil, errors.New("missing `recipients` in SES event")
	}

	event := Event{
		To:    rawEvent.Receipt.Recipients,
		email: []byte(rawEvent.Content),
	}
	return &event, nil
}

var headerToExp = regexp.MustCompile("(^|\n)To: [^\r\n]*(\r?\n)")

// Forward will try to forward the SES event to the given recipient
func (e *Event) Forward(session client.ConfigProvider, to []string) error {
	sesClient := ses.New(session)
	_, err := sesClient.SendRawEmail(&ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{Data: generateMail(e.email, e.To, to)},
	})
	return err
}

func generateMail(raw []byte, originalTo, to []string) []byte {
	buggyRegexBS := " \r\n::OMG::\r\n "
	raw = headerToExp.ReplaceAll(raw, []byte(
		fmt.Sprintf(
			"$1%sTo: %s\r\nX-Original-To: %s$2",
			buggyRegexBS,
			strings.Join(to, ", "),
			strings.Join(originalTo, ", "),
		),
	))
	raw = bytes.Replace(raw, []byte(buggyRegexBS), []byte{}, -1)
	return raw
}
