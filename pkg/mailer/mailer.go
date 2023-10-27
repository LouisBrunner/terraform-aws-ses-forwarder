package mailer

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ses"
)

const passVerdict = "PASS"

type verdict struct {
	Status string `json:"status"`
}

type header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type raw struct {
	Mail struct {
		Headers []header `json:"headers"`
	} `json:"mail"`
	Receipt struct {
		SpamVerdict  verdict `json:"spamVerdict"`
		VirusVerdict verdict `json:"virusVerdict"`
	}
	Content string `json:"content"`
}

// Event contains a SES event
type Event struct {
	From  string
	To    string
	email []byte
}

func lookupHeader(headers []header, name string) (string, bool) {
	for _, header := range headers {
		if header.Name == name {
			return header.Value, true
		}
	}
	return "", false
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
	to, has := lookupHeader(rawEvent.Mail.Headers, "To")
	if !has || to == "" {
		return nil, errors.New("missing `mail.headers.to` in SES event")
	}
	if rawEvent.Receipt.SpamVerdict.Status != passVerdict || rawEvent.Receipt.VirusVerdict.Status != passVerdict {
		return nil, errors.New("don't forward spam/virus")
	}
	from, _ := lookupHeader(rawEvent.Mail.Headers, "From")

	event := Event{
		From:  from,
		To:    to,
		email: []byte(rawEvent.Content),
	}
	return &event, nil
}

var headerFromExp = regexp.MustCompile("(^|\n)From: ([^\r\n]*)(\r?\n)")
var headerToExp = regexp.MustCompile("(^|\n)To: [^\r\n]*(\r?\n)")
var headerReplyToExp = regexp.MustCompile("(^|\n)Reply-To: [^\r\n]*\r?\n")
var headerSenderExp = regexp.MustCompile("(^|\n)Sender: [^\r\n]*\r?\n")
var headerReturnPathExp = regexp.MustCompile("(^|\n)Return-Path: [^\r\n]*\r?\n")
var headerDKIMSigExp = regexp.MustCompile("(^|\n)DKIM-Signature: [^\r\n]*\r?\n")
var headerEndExp = regexp.MustCompile("((\r?\n){2})")

// Forward will try to forward the SES event to the given recipient
func (e *Event) Forward(session client.ConfigProvider, to string) error {
	sesClient := ses.New(session)
	_, err := sesClient.SendRawEmail(&ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{Data: generateMail(e.email, e.To, to)},
	})
	return err
}

func generateMail(raw []byte, originalTo, to string) []byte {
	fromMatches := headerFromExp.FindSubmatch(raw)
	if len(fromMatches) > 3 {
		from := fromMatches[2]
		ending := fromMatches[3]
		addHeader := func(in []byte, name string) []byte {
			locs := headerEndExp.FindIndex(in)

			outStart := in
			outEnd := []byte{}
			if len(locs) > 0 {
				outStart = in[:locs[0]]
				outEnd = in[locs[0]:]
			}
			header := append([]byte{}, ending...)
			header = append(header, append([]byte(name+": "), from...)...)
			result := append([]byte{}, outStart...)
			return append(result, append(header, outEnd...)...)
		}

		if !headerReplyToExp.Match(raw) {
			raw = addHeader(raw, "Reply-To")
		}
		raw = addHeader(raw, "X-Actual-From")
	}

	buggyRegexBS := " \r\n::OMG::\r\n "

	raw = headerFromExp.ReplaceAll(raw, []byte("$1"+buggyRegexBS+"From: "+originalTo+"$3"))
	raw = headerToExp.ReplaceAll(raw, []byte("$1"+buggyRegexBS+"To: "+to+"$2"))
	raw = headerReturnPathExp.ReplaceAll(raw, []byte("$1"))
	raw = headerSenderExp.ReplaceAll(raw, []byte("$1"))
	raw = headerDKIMSigExp.ReplaceAll(raw, []byte("$1"))

	raw = bytes.Replace(raw, []byte(buggyRegexBS), []byte{}, -1)

	return raw
}
