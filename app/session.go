package app

import (
	"fmt"
	"mainichi/adapters"
	"mainichi/core"
	"os"
	"time"
)

type Session struct {
	Store  adapters.Store
	Config core.Config
	Entry  core.Entry

	OpenError   error
	pendingDeck *core.Deck
}

func NewSession(store adapters.Store, cfg core.Config) *Session {
	return &Session{Store: store, Config: cfg}
}

func (s *Session) OpenToday() error {
	return s.OpenDate(time.Now().Format("2006-01-02"))
}

func (s *Session) OpenDate(date string) error {
	s.OpenError = nil
	s.pendingDeck = nil
	entry, err := s.Store.LoadEntry(date)
	if err != nil {
		if !os.IsNotExist(err) {
			s.OpenError = err
			s.Entry = core.Entry{Date: date, Minimum: s.Config.Minimum}
			return err
		}
		entry = core.Entry{
			Date:    date,
			Minimum: s.Config.Minimum,
		}
	}
	if entry.Minimum <= 0 {
		entry.Minimum = s.Config.Minimum
	}
	s.Entry = entry
	return nil
}

func (s *Session) Save() error {
	if s.OpenError != nil {
		return fmt.Errorf("cannot save while entry failed to open: %w", s.OpenError)
	}
	if s.Entry.Minimum <= 0 {
		s.Entry.Minimum = s.Config.Minimum
	}
	if err := s.Store.SaveEntry(s.Entry); err != nil {
		return err
	}
	if s.pendingDeck != nil {
		if err := s.Store.SaveDeckState(*s.pendingDeck); err != nil {
			return err
		}
		s.pendingDeck = nil
	}
	return nil
}

// DrawPrompt loads the deck, draws a prompt, and stages deck state to persist on Save.
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
	core.NormalizeDeck(&deck, prompts)

	prompt := core.Draw(&deck)
	s.Entry.Prompt = prompt
	s.pendingDeck = &deck

	return nil
}

// SetStoicPrompt sets the entry's prompt to the Daily Stoic heading for the entry's date.
// If the entry already has a prompt, it keeps it.
func (s *Session) SetStoicPrompt(headingsJSON string) {
	if s.Entry.Prompt != "" {
		return
	}
	s.Entry.Prompt = core.StoicPrompt(headingsJSON, s.Entry.Date)
}

// ApplyPromptSource assigns a prompt from the configured source.
func (s *Session) ApplyPromptSource(source, defaultPrompts, stoicHeadings string) error {
	if s.Entry.Prompt != "" {
		return nil
	}

	switch source {
	case "stoic":
		s.SetStoicPrompt(stoicHeadings)
		return nil
	default:
		return s.DrawPrompt(defaultPrompts)
	}
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

type CalendarEntry struct {
	WordCount int
	Minimum   int
}

// ListEntries returns a map of day -> calendar entry for entries in the given month.
func (s *Session) ListEntries(year int, month time.Month) map[int]CalendarEntry {
	result := make(map[int]CalendarEntry)
	days := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()

	for day := 1; day <= days; day++ {
		date := fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
		entry, err := s.Store.LoadEntry(date)
		if err != nil {
			continue
		}
		minimum := entry.Minimum
		if minimum <= 0 {
			minimum = s.Config.Minimum
		}
		result[day] = CalendarEntry{
			WordCount: core.WordCount(entry.Body),
			Minimum:   minimum,
		}
	}
	return result
}
