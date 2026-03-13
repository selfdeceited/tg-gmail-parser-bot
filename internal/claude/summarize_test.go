package claude

import (
	"encoding/json"
	"testing"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/gmail"
)

// ---------- stripCodeFence ----------

func Test_stripCodeFence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no fence — plain JSON",
			input: `{"result":"matched"}`,
			want:  `{"result":"matched"}`,
		},
		{
			name:  "json fence",
			input: "```json\n{\"result\":\"matched\"}\n```",
			want:  `{"result":"matched"}`,
		},
		{
			name:  "generic fence",
			input: "```\n{\"result\":\"not matched\"}\n```",
			want:  `{"result":"not matched"}`,
		},
		{
			name:  "fence with leading/trailing whitespace",
			input: "  ```json\n{\"ok\":true}\n```  ",
			want:  `{"ok":true}`,
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only backticks",
			input: "```",
			want:  "",
		},
		{
			name:  "fence no closing",
			input: "```json\n{\"result\":\"matched\"}",
			want:  `{"result":"matched"}`,
		},
		{
			name:  "multiple closing fences — trim at last",
			input: "```json\nfoo```bar```",
			want:  "foo```bar",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := stripCodeFence(tc.input)
			if got != tc.want {
				t.Errorf("stripCodeFence(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ---------- buildPrompt ----------

func Test_buildPrompt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		userPrompt     string
		email          gmail.EmailMessage
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:       "all fields present",
			userPrompt: "Alert me about invoices",
			email: gmail.EmailMessage{
				Subject: "Invoice #1234",
				From:    "billing@acme.com",
				Body:    "Your invoice is ready.",
			},
			wantContains: []string{
				"Alert me about invoices",
				"Invoice #1234",
				"billing@acme.com",
				"Your invoice is ready.",
				`"matched" or "not matched"`,
			},
		},
		{
			name:       "empty body and from",
			userPrompt: "Any email",
			email: gmail.EmailMessage{
				Subject: "Hello",
				From:    "",
				Body:    "",
			},
			wantContains: []string{
				"Any email",
				"Hello",
			},
		},
		{
			name:           "prompt injected into output once",
			userPrompt:     "UNIQUE_PROMPT_STRING",
			email:          gmail.EmailMessage{},
			wantContains:   []string{"UNIQUE_PROMPT_STRING"},
			wantNotContain: []string{"UNIQUE_PROMPT_STRING\nUNIQUE_PROMPT_STRING"}, // not duplicated
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := buildPrompt(tc.userPrompt, tc.email)
			for _, sub := range tc.wantContains {
				if !contains(got, sub) {
					t.Errorf("buildPrompt output missing %q\ngot: %s", sub, got)
				}
			}
			for _, sub := range tc.wantNotContain {
				if contains(got, sub) {
					t.Errorf("buildPrompt output should not contain %q\ngot: %s", sub, got)
				}
			}
		})
	}
}

// ---------- SummarizeResult.ContentString ----------

func TestSummarizeResult_ContentString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content json.RawMessage
		want    string
	}{
		{
			name:    "JSON string content",
			content: json.RawMessage(`"This is a plain summary"`),
			want:    "This is a plain summary",
		},
		{
			name:    "JSON object content — pretty printed",
			content: json.RawMessage(`{"key":"value"}`),
			want:    "{\n  \"key\": \"value\"\n}",
		},
		{
			name:    "JSON array content — pretty printed",
			content: json.RawMessage(`[1,2,3]`),
			want:    "[\n  1,\n  2,\n  3\n]",
		},
		{
			name:    "nil content",
			content: nil,
			want:    "",
		},
		{
			name:    "empty raw message",
			content: json.RawMessage{},
			want:    "",
		},
		{
			name:    "invalid JSON falls back to raw bytes",
			content: json.RawMessage(`not-json`),
			want:    "not-json",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := &SummarizeResult{Content: tc.content}
			got := r.ContentString()
			if got != tc.want {
				t.Errorf("ContentString() = %q, want %q", got, tc.want)
			}
		})
	}
}

// ---------- helpers ----------

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexStr(s, sub) >= 0)
}

func indexStr(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
