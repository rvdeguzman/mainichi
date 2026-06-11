package core

import (
	"math/rand"
	"slices"
	"strings"
)

type Deck struct {
	Prompts   []string `json:"prompts"`
	Remaining []int    `json:"remaining"`
	Used      []int    `json:"used"`
}

func LoadPrompts(lines []string) []string {
	var prompts []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			prompts = append(prompts, line)
		}
	}
	return prompts
}

func NewDeck(prompts []string) Deck {
	d := Deck{Prompts: prompts}
	Reshuffle(&d)
	return d
}

func NormalizeDeck(deck *Deck, prompts []string) {
	if !slices.Equal(deck.Prompts, prompts) {
		*deck = NewDeck(prompts)
		return
	}
	for _, idx := range deck.Remaining {
		if idx < 0 || idx >= len(deck.Prompts) {
			Reshuffle(deck)
			return
		}
	}
	for _, idx := range deck.Used {
		if idx < 0 || idx >= len(deck.Prompts) {
			Reshuffle(deck)
			return
		}
	}
}

func Draw(deck *Deck) string {
	if len(deck.Prompts) == 0 {
		return ""
	}
	NormalizeDeck(deck, deck.Prompts)
	if NeedsReshuffle(*deck) {
		Reshuffle(deck)
	}
	idx := deck.Remaining[0]
	deck.Remaining = deck.Remaining[1:]
	deck.Used = append(deck.Used, idx)
	return deck.Prompts[idx]
}

func Reshuffle(deck *Deck) {
	deck.Remaining = make([]int, len(deck.Prompts))
	for i := range deck.Remaining {
		deck.Remaining[i] = i
	}
	rand.Shuffle(len(deck.Remaining), func(i, j int) {
		deck.Remaining[i], deck.Remaining[j] = deck.Remaining[j], deck.Remaining[i]
	})
	deck.Used = nil
}

func NeedsReshuffle(deck Deck) bool {
	return len(deck.Remaining) == 0
}
