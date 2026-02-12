package core

import (
	"testing"
)

func TestDeckExhaustion(t *testing.T) {
	prompts := []string{"one", "two", "three"}
	deck := NewDeck(prompts)

	seen := make(map[string]bool)
	for range prompts {
		p := Draw(&deck)
		if seen[p] {
			t.Errorf("got duplicate prompt %q before deck exhaustion", p)
		}
		seen[p] = true
	}

	if len(seen) != len(prompts) {
		t.Errorf("expected %d unique prompts, got %d", len(prompts), len(seen))
	}
}

func TestDeckReshuffle(t *testing.T) {
	prompts := []string{"a", "b"}
	deck := NewDeck(prompts)

	// Exhaust
	Draw(&deck)
	Draw(&deck)

	if !NeedsReshuffle(deck) {
		t.Error("deck should need reshuffle after exhaustion")
	}

	// Drawing again should trigger reshuffle
	p := Draw(&deck)
	if p != "a" && p != "b" {
		t.Errorf("unexpected prompt after reshuffle: %q", p)
	}
}

func TestLoadPrompts(t *testing.T) {
	lines := []string{"  hello  ", "", "world", "  "}
	got := LoadPrompts(lines)
	if len(got) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(got))
	}
}
