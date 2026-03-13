package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/claude"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/gmail"
)

// ---------- filterPrompts ----------

func Test_filterPrompts(t *testing.T) {
	t.Parallel()

	promptWithFilter := func(id, filter string) entities.Prompt {
		return entities.Prompt{ID: uuid.MustParse("00000000-0000-0000-0000-" + strings.Repeat("0", 11) + id), Filter: filter, Prompt: "p" + id}
	}
	promptNoFilter := func(id string) entities.Prompt {
		return entities.Prompt{ID: uuid.MustParse("00000000-0000-0000-0000-" + strings.Repeat("0", 11) + id), Filter: "", Prompt: "p" + id}
	}

	tests := []struct {
		name        string
		email       gmail.EmailMessage
		prompts     []entities.Prompt
		wantFilters []string // expected Filter values of returned prompts
		wantLen     int
	}{
		{
			name:    "no prompts — empty result",
			email:   gmail.EmailMessage{From: "x@example.com"},
			prompts: []entities.Prompt{},
			wantLen: 0,
		},
		{
			name:  "only no-filter prompts — all returned",
			email: gmail.EmailMessage{From: "x@example.com"},
			prompts: []entities.Prompt{
				promptNoFilter("000000000001"),
				promptNoFilter("000000000002"),
			},
			wantLen: 2,
		},
		{
			name:  "sender matches filter — only that prompt returned",
			email: gmail.EmailMessage{From: "billing@acme.com"},
			prompts: []entities.Prompt{
				promptNoFilter("000000000001"),
				promptWithFilter("000000000002", "billing@acme.com"),
				promptNoFilter("000000000003"),
			},
			wantFilters: []string{"billing@acme.com"},
			wantLen:     1,
		},
		{
			name:  "filter match is case-insensitive",
			email: gmail.EmailMessage{From: "BILLING@ACME.COM"},
			prompts: []entities.Prompt{
				promptWithFilter("000000000001", "billing@acme.com"),
			},
			wantFilters: []string{"billing@acme.com"},
			wantLen:     1,
		},
		{
			name:  "no filter match — only no-filter prompts returned",
			email: gmail.EmailMessage{From: "unknown@other.com"},
			prompts: []entities.Prompt{
				promptWithFilter("000000000001", "billing@acme.com"),
				promptNoFilter("000000000002"),
			},
			wantFilters: []string{""},
			wantLen:     1,
		},
		{
			name:  "filter match wins over no-filter prompts",
			email: gmail.EmailMessage{From: "alerts@service.io"},
			prompts: []entities.Prompt{
				promptNoFilter("000000000001"),
				promptWithFilter("000000000002", "alerts@service.io"),
			},
			wantFilters: []string{"alerts@service.io"},
			wantLen:     1,
		},
		{
			name:  "multiple no-filter prompts, no match — all no-filter returned",
			email: gmail.EmailMessage{From: "noreply@test.com"},
			prompts: []entities.Prompt{
				promptNoFilter("000000000001"),
				promptNoFilter("000000000002"),
				promptWithFilter("000000000003", "other@example.com"),
			},
			wantLen: 2,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := filterPrompts(tc.email, tc.prompts)
			if len(got) != tc.wantLen {
				t.Errorf("len(filterPrompts) = %d, want %d; got %+v", len(got), tc.wantLen, got)
			}
			if tc.wantFilters != nil {
				for i, wf := range tc.wantFilters {
					if i >= len(got) {
						t.Errorf("missing prompt at index %d", i)
						continue
					}
					if got[i].Filter != wf {
						t.Errorf("got[%d].Filter = %q, want %q", i, got[i].Filter, wf)
					}
				}
			}
		})
	}
}

// ---------- formatSummary ----------

func Test_formatSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		result        *claude.SummarizeResult
		url           string
		promptShortID string
		wantContains  []string
	}{
		{
			name: "basic summary",
			result: &claude.SummarizeResult{
				Result:  "matched",
				Title:   "Invoice #1234",
				Content: json.RawMessage(`"Your invoice is due."`),
			},
			url:           "https://mail.google.com/mail/u/0/#all/abc123",
			promptShortID: "abc123",
			wantContains: []string{
				"Invoice #1234",
				"Your invoice is due.",
				"https://mail.google.com/mail/u/0/#all/abc123",
				"abc123",
				"Open in Gmail",
			},
		},
		{
			name: "HTML special chars in title are escaped",
			result: &claude.SummarizeResult{
				Result:  "matched",
				Title:   "<script>alert('xss')</script>",
				Content: json.RawMessage(`"safe content"`),
			},
			url:           "https://mail.google.com/mail/u/0/#all/id1",
			promptShortID: "id1",
			wantContains: []string{
				"&lt;script&gt;",
				"safe content",
			},
		},
		{
			name: "empty content",
			result: &claude.SummarizeResult{
				Result:  "matched",
				Title:   "Empty",
				Content: json.RawMessage(`""`),
			},
			url:           "https://mail.google.com/mail/u/0/#all/id2",
			promptShortID: "id2",
			wantContains:  []string{"Empty"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := formatSummary(tc.result, tc.url, tc.promptShortID)
			for _, sub := range tc.wantContains {
				if !strings.Contains(got, sub) {
					t.Errorf("formatSummary output missing %q\ngot: %s", sub, got)
				}
			}
		})
	}
}
