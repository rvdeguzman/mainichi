package main

import (
	"strings"
	"testing"
)

func TestNormalizeCommandSupportsAIAliases(t *testing.T) {
	tests := map[string]string{
		"ai":     "ai",
		"--ai":   "ai",
		"prompt": "prompt",
	}

	for input, want := range tests {
		if got := normalizeCommand(input); got != want {
			t.Fatalf("normalizeCommand(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestUsageDocumentsAIAliases(t *testing.T) {
	var b strings.Builder
	printUsage(&b)

	got := b.String()
	if !strings.Contains(got, "ai, --ai") {
		t.Fatalf("usage = %q, want AI command aliases", got)
	}
}
