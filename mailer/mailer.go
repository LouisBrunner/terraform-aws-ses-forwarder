package mailer

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ses"
)

const passVerdict = "PASS"

type verdict struct {
	Status string `json:"status"`
}

type raw struct {
	Mail struct {
		Headers struct {
			To string `json:"to"`
		} `json:"headers"`
	} `json:"mail"`
	Receipt struct {
		SpamVerdict  verdict `json:"spamVerdict"`
		VirusVerdict verdict `json:"virusVerdict"`
	}
	Content string `json:"content"`
}

// Event contains a SES event
type Event struct {
	To    string
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
	if rawEvent.Mail.Headers.To == "" {
		return nil, errors.New("missing `mail.headers.to` in SES event")
	}
	if rawEvent.Receipt.SpamVerdict.Status != passVerdict || rawEvent.Receipt.VirusVerdict.Status != passVerdict {
		return nil, errors.New("don't forward spam/virus")
	}

	event := Event{
		To:    rawEvent.Mail.Headers.To,
		email: []byte(rawEvent.Content),
	}
	return &event, nil
}

// Forward will try to forward the SES event to the given recipient
func (e *Event) Forward(session client.ConfigProvider, to string) error {
	sesClient := ses.New(session)

	raw := bytes.Replace(
		e.email,
		[]byte("To: "+to+"\r\n"),
		[]byte("To: "+e.To+"\r\n"),
		-1,
	)

	_, err := sesClient.SendRawEmail(&ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{Data: raw},
	})
	return err
}
