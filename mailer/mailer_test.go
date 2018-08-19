package mailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const eventValid = `{
  "content":"123",
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"}},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`

func TestParseEvent_Works(t *testing.T) {
	event, err := ParseEvent([]byte(eventValid))
	if assert.NoError(t, err) {
		assert.Equal(t, "hello@moto.com", event.To)
		assert.Equal(t, []byte("123"), event.email)
	}
}

const eventInvalid = `{
  "content":"123",
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"}},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]},
}`

func TestParseEvent_Fails_Invalid(t *testing.T) {
	_, err := ParseEvent([]byte(eventInvalid))
	assert.EqualError(t, err, "invalid character '}' looking for beginning of object key string")
}

const eventMissingContent = `{
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"}},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`

func TestParseEvent_Fails_MissingContent(t *testing.T) {
	_, err := ParseEvent([]byte(eventMissingContent))
	assert.EqualError(t, err, "missing `content` in SES event")
}

const eventMissingTo = `{
  "content":"123",
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"}},
  "mail":{"headers":[{"name":"From","value":"hello@moto.com"}]}
}`

func TestParseEvent_Fails_MissingTo(t *testing.T) {
	_, err := ParseEvent([]byte(eventMissingTo))
	assert.EqualError(t, err, "missing `mail.headers.to` in SES event")
}

const eventIsSpam = `{
  "content":"123",
  "receipt":{"spamVerdict":{"status":"FAIL"},"virusVerdict":{"status":"PASS"}},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`

func TestParseEvent_Fails_IsSpam(t *testing.T) {
	_, err := ParseEvent([]byte(eventIsSpam))
	assert.EqualError(t, err, "don't forward spam/virus")
}

const eventIsVirus = `{
  "content":"123",
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"FAIL"}},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`

func TestParseEvent_Fails_IsVirus(t *testing.T) {
	_, err := ParseEvent([]byte(eventIsVirus))
	assert.EqualError(t, err, "don't forward spam/virus")
}

const rawEmail = "From: src <src@mail>\r\nDKIM-Signature: efewfew\r\nTo: dest <dest@mail>\r\nReturn-Path: hey\r\nSender: wazzup\r\n\r\nCONTENT"

func TestGenerateEmil_Works(t *testing.T) {
	result := generateMail([]byte(rawEmail), "dest <dest@mail>", "new@mail")
	assert.Equal(t, "From: dest <dest@mail>\r\nTo: new@mail\r\nReply-To: src <src@mail>\r\nX-Actual-From: src <src@mail>\r\n\r\nCONTENT", string(result))
}

const rawEmailNoFrom = "DKIM-Signature: efewfew\r\nTo: dest <dest@mail>\r\nDKIM-Signature: efewfew\r\n\r\nCONTENT"

func TestGenerateEmil_Works_NoFrom(t *testing.T) {
	result := generateMail([]byte(rawEmailNoFrom), "dest <dest@mail>", "new@mail")
	assert.Equal(t, "To: new@mail\r\n\r\nCONTENT", string(result))
}

const rawEmailReplyTwo = "From: src <src@mail>\r\nDKIM-Signature: efewfew\r\nTo: dest <dest@mail>\r\nReply-To: hey\r\nDKIM-Signature: efewfew\n\nCONTENT"

func TestGenerateEmil_Works_WithReplyTo(t *testing.T) {
	result := generateMail([]byte(rawEmailReplyTwo), "dest <dest@mail>", "new@mail")
	assert.Equal(t, "From: dest <dest@mail>\r\nTo: new@mail\r\nReply-To: hey\r\nX-Actual-From: src <src@mail>\n\nCONTENT", string(result))
}
