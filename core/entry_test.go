package core

import (
	"testing"
)

func TestRoundTrip(t *testing.T) {
	original := Entry{
		Date:    "2026-02-12",
		Prompt:  "Where are you mistaking motion for progress?",
		Minimum: 300,
		Body:    "I woke up today thinking about clarity.\n",
	}

	data := SerializeEntry(original)
	parsed, err := ParseEntry(data)
	if err != nil {
		t.Fatalf("ParseEntry failed: %v", err)
	}

	if parsed.Date != original.Date {
		t.Errorf("Date: got %q, want %q", parsed.Date, original.Date)
	}
	if parsed.Prompt != original.Prompt {
		t.Errorf("Prompt: got %q, want %q", parsed.Prompt, original.Prompt)
	}
	if parsed.Minimum != original.Minimum {
		t.Errorf("Minimum: got %d, want %d", parsed.Minimum, original.Minimum)
	}
	if parsed.Body != original.Body {
		t.Errorf("Body: got %q, want %q", parsed.Body, original.Body)
	}
}

func TestRoundTripNoPrompt(t *testing.T) {
	original := Entry{
		Date:    "2026-01-15",
		Minimum: 300,
		Body:    "Just writing.\n",
	}

	data := SerializeEntry(original)
	parsed, err := ParseEntry(data)
	if err != nil {
		t.Fatalf("ParseEntry failed: %v", err)
	}

	if parsed.Prompt != "" {
		t.Errorf("Prompt: got %q, want empty", parsed.Prompt)
	}
	if parsed.Body != original.Body {
		t.Errorf("Body: got %q, want %q", parsed.Body, original.Body)
	}
}

func TestWordCount(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"  spaces   everywhere  ", 2},
		{"one\ntwo\nthree", 3},
		{"tabs\there\ttoo", 3},
	}

	for _, tt := range tests {
		got := WordCount(tt.text)
		if got != tt.want {
			t.Errorf("WordCount(%q) = %d, want %d", tt.text, got, tt.want)
		}
	}
}
