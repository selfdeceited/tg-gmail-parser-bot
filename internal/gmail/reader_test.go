package gmail

import (
	"encoding/base64"
	"testing"

	gmailapi "google.golang.org/api/gmail/v1"
)

// ---------- extractBody ----------

func Test_extractBody(t *testing.T) {
	t.Parallel()

	encode := func(s string) string {
		return base64.URLEncoding.EncodeToString([]byte(s))
	}

	tests := []struct {
		name string
		part *gmailapi.MessagePart
		want string
	}{
		{
			name: "nil part",
			part: nil,
			want: "",
		},
		{
			name: "text/plain at top level",
			part: &gmailapi.MessagePart{
				MimeType: "text/plain",
				Body:     &gmailapi.MessagePartBody{Data: encode("Hello, World!")},
			},
			want: "Hello, World!",
		},
		{
			name: "text/plain; charset=utf-8",
			part: &gmailapi.MessagePart{
				MimeType: "text/plain; charset=utf-8",
				Body:     &gmailapi.MessagePartBody{Data: encode("UTF-8 content")},
			},
			want: "UTF-8 content",
		},
		{
			name: "text/html — not extracted",
			part: &gmailapi.MessagePart{
				MimeType: "text/html",
				Body:     &gmailapi.MessagePartBody{Data: encode("<p>hello</p>")},
			},
			want: "",
		},
		{
			name: "multipart/alternative with text/plain child",
			part: &gmailapi.MessagePart{
				MimeType: "multipart/alternative",
				Body:     &gmailapi.MessagePartBody{},
				Parts: []*gmailapi.MessagePart{
					{
						MimeType: "text/html",
						Body:     &gmailapi.MessagePartBody{Data: encode("<p>html</p>")},
					},
					{
						MimeType: "text/plain",
						Body:     &gmailapi.MessagePartBody{Data: encode("plain text")},
					},
				},
			},
			want: "plain text",
		},
		{
			name: "nested multipart",
			part: &gmailapi.MessagePart{
				MimeType: "multipart/mixed",
				Body:     &gmailapi.MessagePartBody{},
				Parts: []*gmailapi.MessagePart{
					{
						MimeType: "multipart/alternative",
						Body:     &gmailapi.MessagePartBody{},
						Parts: []*gmailapi.MessagePart{
							{
								MimeType: "text/plain",
								Body:     &gmailapi.MessagePartBody{Data: encode("nested plain")},
							},
						},
					},
				},
			},
			want: "nested plain",
		},
		{
			name: "text/plain with empty body data",
			part: &gmailapi.MessagePart{
				MimeType: "text/plain",
				Body:     &gmailapi.MessagePartBody{Data: ""},
			},
			want: "",
		},
		{
			name: "text/plain with invalid base64",
			part: &gmailapi.MessagePart{
				MimeType: "text/plain",
				Body:     &gmailapi.MessagePartBody{Data: "!!!not-base64!!!"},
			},
			want: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := extractBody(tc.part)
			if got != tc.want {
				t.Errorf("extractBody() = %q, want %q", got, tc.want)
			}
		})
	}
}

// ---------- parseMessage ----------

func Test_parseMessage(t *testing.T) {
	t.Parallel()

	encode := func(s string) string {
		return base64.URLEncoding.EncodeToString([]byte(s))
	}

	tests := []struct {
		name         string
		msg          *gmailapi.Message
		accountIndex int
		want         EmailMessage
	}{
		{
			name: "basic message — headers and snippet",
			msg: &gmailapi.Message{
				Id:      "msg001",
				Snippet: "snippet text",
				Payload: &gmailapi.MessagePart{
					MimeType: "text/plain",
					Headers: []*gmailapi.MessagePartHeader{
						{Name: "Subject", Value: "Test Subject"},
						{Name: "From", Value: "sender@example.com"},
					},
					Body: &gmailapi.MessagePartBody{Data: ""},
				},
			},
			accountIndex: 0,
			want: EmailMessage{
				ID:      "msg001",
				Subject: "Test Subject",
				From:    "sender@example.com",
				Body:    "snippet text",
				URL:     "https://mail.google.com/mail/u/0/#all/msg001",
			},
		},
		{
			name: "body overrides snippet when text/plain present",
			msg: &gmailapi.Message{
				Id:      "msg002",
				Snippet: "short snippet",
				Payload: &gmailapi.MessagePart{
					MimeType: "text/plain",
					Headers: []*gmailapi.MessagePartHeader{
						{Name: "Subject", Value: "Rich Body"},
						{Name: "From", Value: "a@b.com"},
					},
					Body: &gmailapi.MessagePartBody{Data: encode("Full body text here")},
				},
			},
			accountIndex: 1,
			want: EmailMessage{
				ID:      "msg002",
				Subject: "Rich Body",
				From:    "a@b.com",
				Body:    "Full body text here",
				URL:     "https://mail.google.com/mail/u/1/#all/msg002",
			},
		},
		{
			name: "case-insensitive header matching",
			msg: &gmailapi.Message{
				Id:      "msg003",
				Snippet: "",
				Payload: &gmailapi.MessagePart{
					MimeType: "text/plain",
					Headers: []*gmailapi.MessagePartHeader{
						{Name: "SUBJECT", Value: "Loud Subject"},
						{Name: "FROM", Value: "loud@sender.com"},
					},
					Body: &gmailapi.MessagePartBody{Data: ""},
				},
			},
			accountIndex: 2,
			want: EmailMessage{
				ID:      "msg003",
				Subject: "Loud Subject",
				From:    "loud@sender.com",
				Body:    "",
				URL:     "https://mail.google.com/mail/u/2/#all/msg003",
			},
		},
		{
			name: "account index encoded in URL",
			msg: &gmailapi.Message{
				Id:      "msg004",
				Snippet: "",
				Payload: &gmailapi.MessagePart{
					MimeType: "text/plain",
					Headers:  []*gmailapi.MessagePartHeader{},
					Body:     &gmailapi.MessagePartBody{Data: ""},
				},
			},
			accountIndex: 5,
			want: EmailMessage{
				ID:  "msg004",
				URL: "https://mail.google.com/mail/u/5/#all/msg004",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := parseMessage(tc.msg, tc.accountIndex)
			if got.ID != tc.want.ID {
				t.Errorf("ID = %q, want %q", got.ID, tc.want.ID)
			}
			if got.Subject != tc.want.Subject {
				t.Errorf("Subject = %q, want %q", got.Subject, tc.want.Subject)
			}
			if got.From != tc.want.From {
				t.Errorf("From = %q, want %q", got.From, tc.want.From)
			}
			if got.Body != tc.want.Body {
				t.Errorf("Body = %q, want %q", got.Body, tc.want.Body)
			}
			if got.URL != tc.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tc.want.URL)
			}
		})
	}
}
