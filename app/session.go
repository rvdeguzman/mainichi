package app

import (
	"fmt"
	"mainichi/adapters"
	"mainichi/core"
	"time"
)

type Session struct {
	Store  adapters.Store
	Config core.Config
	Entry  core.Entry
}

func NewSession(store adapters.Store, cfg core.Config) *Session {
	return &Session{Store: store, Config: cfg}
}

func (s *Session) OpenToday() {
	s.OpenDate(time.Now().Format("2006-01-02"))
}

func (s *Session) OpenDate(date string) {
	entry, err := s.Store.LoadEntry(date)
	if err != nil {
		entry = core.Entry{
			Date:    date,
			Minimum: s.Config.Minimum,
		}
	}
	s.Entry = entry
}

func (s *Session) Save() error {
	s.Entry.Minimum = s.Config.Minimum
	return s.Store.SaveEntry(s.Entry)
}

// DrawPrompt loads the deck, draws a prompt, assigns it to the entry, and persists deck state.
// If the entry already has a prompt, it keeps it.
func (s *Session) DrawPrompt(defaultPrompts string) error {
	if s.Entry.Prompt != "" {
		return nil
	}

	// Ensure prompts.txt exists
	if err := s.Store.EnsureDefaultPrompts(defaultPrompts); err != nil {
		return err
	}

	prompts, err := s.Store.LoadPrompts()
	if err != nil {
		return err
	}

	deck, err := s.Store.LoadDeckState()
	if err != nil {
		// No saved state — create a fresh deck
		deck = core.NewDeck(prompts)
	}
	// Sync prompts list (in case prompts.txt changed)
	deck.Prompts = prompts

	prompt := core.Draw(&deck)
	s.Entry.Prompt = prompt

	return s.Store.SaveDeckState(deck)
}

type RecentEntry struct {
	Entry     core.Entry
	WordCount int
}

func (s *Session) ListRecentEntries(limit int) ([]RecentEntry, error) {
	dates, err := s.Store.ListEntryDates(limit)
	if err != nil {
		return nil, err
	}

	var result []RecentEntry
	for _, date := range dates {
		entry, err := s.Store.LoadEntry(date)
		if err != nil {
			continue
		}
		result = append(result, RecentEntry{
			Entry:     entry,
			WordCount: core.WordCount(entry.Body),
		})
	}
	return result, nil
}

// ListEntries returns a map of day -> word count for entries in the given month.
func (s *Session) ListEntries(year int, month time.Month) map[int]int {
	result := make(map[int]int)
	days := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()

	for day := 1; day <= days; day++ {
		date := fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
		entry, err := s.Store.LoadEntry(date)
		if err != nil {
			continue
		}
		result[day] = core.WordCount(entry.Body)
	}
	return result
}
