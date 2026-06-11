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

func TestDrawResetsStaleOutOfRangeIndexes(t *testing.T) {
	deck := Deck{Prompts: []string{"only"}, Remaining: []int{4}, Used: []int{0}}

	got := Draw(&deck)
	if got != "only" {
		t.Fatalf("Draw() = %q, want only", got)
	}
}

func TestNormalizeDeckResetsWhenPromptsChanged(t *testing.T) {
	deck := Deck{Prompts: []string{"old-a", "old-b"}, Remaining: []int{1}, Used: []int{0}}
	NormalizeDeck(&deck, []string{"new-only"})

	if len(deck.Prompts) != 1 || deck.Prompts[0] != "new-only" {
		t.Fatalf("Prompts = %#v, want new-only", deck.Prompts)
	}
	if len(deck.Remaining) != 1 || deck.Remaining[0] != 0 {
		t.Fatalf("Remaining = %#v, want [0]", deck.Remaining)
	}
	if len(deck.Used) != 0 {
		t.Fatalf("Used = %#v, want reset", deck.Used)
	}
}
