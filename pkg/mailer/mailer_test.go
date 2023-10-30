package mailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEvent(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		content  string
		expected *Event
		wantErr  bool
	}{
		{
			name: "works",
			content: `{
  "content":"MTIz",
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"},"recipients":["hello@moto.com"]},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`,
			expected: &Event{
				To: []string{
					"hello@moto.com",
				},
				email: []byte("123"),
			},
		},
		{
			name: "works (disabled)",
			content: `{
  "content":"MTIz",
  "receipt":{"spamVerdict":{"status":"DISABLED"},"virusVerdict":{"status":"PASS"},"recipients":["hello@moto.com"]},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`,
			expected: &Event{
				To: []string{
					"hello@moto.com",
				},
				email: []byte("123"),
			},
		},
		{
			name: "works (both disabled)",
			content: `{
  "content":"MTIz",
  "receipt":{"spamVerdict":{"status":"DISABLED"},"virusVerdict":{"status":"DISABLED"},"recipients":["hello@moto.com"]},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`,
			expected: &Event{
				To: []string{
					"hello@moto.com",
				},
				email: []byte("123"),
			},
		},
		{
			name: "fails (invalid json)",
			content: `{
  "content":"MTIz",
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"},"recipients":["hello@moto.com"}]},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]},
}`,
			wantErr: true,
		},
		{
			name: "fails (no content)",
			content: `{
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"},"recipients":["hello@moto.com"]},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`,
			wantErr: true,
		},
		{
			name: "fails (no to)",
			content: `{
  "content":"MTIz",
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"}},
  "mail":{"headers":[{"name":"From","value":"hello@moto.com"}]}
}`,
			wantErr: true,
		},
		{
			name: "fails (spam)",
			content: `{
  "content":"MTIz",
  "receipt":{"spamVerdict":{"status":"FAIL"},"virusVerdict":{"status":"PASS"},"recipients":["hello@moto.com"]},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`,
			wantErr: true,
		},
		{
			name: "fails (virus)",
			content: `{
  "content":"MTIz",
  "receipt":{"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"FAIL"},"recipients":["hello@moto.com"]},
  "mail":{"headers":[{"name":"To","value":"hello@moto.com"}]}
}`,
			wantErr: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			got, err := ParseEvent([]byte(testcase.content))
			if testcase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testcase.expected, got)
			}
		})
	}
}

func TestGenerateEmail(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		email      string
		originalTo []string
		to         []string
		expected   string
	}{
		{
			name:       "works",
			email:      "From: src <src@mail>\r\nDKIM-Signature: efewfew\r\nTo: dest <dest@mail>\r\nReturn-Path: hey\r\nSender: wazzup\r\n\r\nCONTENT",
			originalTo: []string{"dest <dest@mail>"},
			to:         []string{"new@mail"},
			expected:   "From: src <src@mail>\r\nDKIM-Signature: efewfew\r\nTo: new@mail\r\nX-Original-To: dest <dest@mail>\r\nReturn-Path: hey\r\nSender: wazzup\r\n\r\nCONTENT",
		},
		{
			name:       "no from",
			email:      "DKIM-Signature: efewfew\r\nTo: dest <dest@mail>\r\nDKIM-Signature: efewfew\r\n\r\nCONTENT",
			originalTo: []string{"dest <dest@mail>"},
			to:         []string{"new@mail"},
			expected:   "DKIM-Signature: efewfew\r\nTo: new@mail\r\nX-Original-To: dest <dest@mail>\r\nDKIM-Signature: efewfew\r\n\r\nCONTENT",
		},
		{
			name:       "reply to",
			email:      "From: src <src@mail>\r\nDKIM-Signature: efewfew\r\nTo: dest <dest@mail>\r\nReply-To: hey\r\nDKIM-Signature: efewfew\n\nCONTENT",
			originalTo: []string{"dest <dest@mail>"},
			to:         []string{"new@mail"},
			expected:   "From: src <src@mail>\r\nDKIM-Signature: efewfew\r\nTo: new@mail\r\nX-Original-To: dest <dest@mail>\r\nReply-To: hey\r\nDKIM-Signature: efewfew\n\nCONTENT",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			got := generateMail([]byte(testcase.email), testcase.originalTo, testcase.to)
			assert.Equal(t, testcase.expected, string(got))
		})
	}
}
