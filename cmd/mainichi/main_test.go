package main

import (
	"strings"
	"testing"
)

func TestSupportedPromptSource(t *testing.T) {
	tests := map[string]string{
		"deck":    "deck",
		"stoic":   "stoic",
		"":        "stoic",
		"unknown": "stoic",
	}

	for input, want := range tests {
		if got := supportedPromptSource(input); got != want {
			t.Fatalf("supportedPromptSource(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestUsageDocumentsCommands(t *testing.T) {
	var b strings.Builder
	printUsage(&b)

	got := b.String()
	for _, command := range []string{"prompt", "stoic", "config", "date", "recent", "YYYY-MM-DD"} {
		if !strings.Contains(got, command) {
			t.Fatalf("usage = %q, want command %q", got, command)
		}
	}
}
