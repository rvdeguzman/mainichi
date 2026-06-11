package app

import (
	"os"
	"path/filepath"
	"testing"

	"mainichi/adapters"
	"mainichi/core"
)

func newTestSession(t *testing.T) (*Session, string) {
	t.Helper()
	dir := t.TempDir()
	store := adapters.NewStore(dir)
	return NewSession(store, core.DefaultConfig()), dir
}

func TestOpenDateMissingEntryStartsNewEntry(t *testing.T) {
	s, _ := newTestSession(t)

	if err := s.OpenDate("2026-06-11"); err != nil {
		t.Fatalf("OpenDate missing entry returned error: %v", err)
	}

	if s.Entry.Date != "2026-06-11" {
		t.Fatalf("Date = %q, want requested date", s.Entry.Date)
	}
	if s.OpenError != nil {
		t.Fatalf("OpenError = %v, want nil", s.OpenError)
	}
}

func TestOpenDateMalformedEntryReturnsErrorAndDoesNotSilentlyReplace(t *testing.T) {
	s, dir := newTestSession(t)
	entries := filepath.Join(dir, "entries")
	if err := os.MkdirAll(entries, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(entries, "2026-06-11.md")
	bad := []byte("not frontmatter\nimportant draft")
	if err := os.WriteFile(path, bad, 0o644); err != nil {
		t.Fatal(err)
	}

	err := s.OpenDate("2026-06-11")
	if err == nil {
		t.Fatal("OpenDate malformed entry error = nil, want error")
	}
	if s.OpenError == nil {
		t.Fatal("OpenError = nil, want visible stored error")
	}
	if s.Entry.Body != "" {
		t.Fatalf("Entry body = %q, want empty safe entry", s.Entry.Body)
	}

	if err := s.Save(); err == nil {
		t.Fatal("Save after failed open error = nil, want refusal to overwrite malformed file")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(bad) {
		t.Fatalf("malformed file was overwritten; got %q", string(got))
	}
}

func TestDeckPromptCommitsOnlyAfterSuccessfulEntrySave(t *testing.T) {
	s, dir := newTestSession(t)
	if err := os.MkdirAll(filepath.Join(dir, "entries"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := s.OpenDate("2026-06-11"); err != nil {
		t.Fatal(err)
	}
	if err := s.DrawPrompt("one\ntwo\n"); err != nil {
		t.Fatalf("DrawPrompt: %v", err)
	}
	if s.Entry.Prompt == "" {
		t.Fatal("DrawPrompt did not assign prompt")
	}
	if _, err := os.Stat(filepath.Join(dir, "prompt_state.json")); !os.IsNotExist(err) {
		t.Fatalf("deck state exists before Save; err=%v", err)
	}

	// Make entries path unwritable by replacing the entries directory with a file.
	if err := os.RemoveAll(filepath.Join(dir, "entries")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "entries"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err == nil {
		t.Fatal("Save error = nil, want entry save failure")
	}
	if _, err := os.Stat(filepath.Join(dir, "prompt_state.json")); !os.IsNotExist(err) {
		t.Fatalf("deck state committed after failed entry save; err=%v", err)
	}

	if err := os.Remove(filepath.Join(dir, "entries")); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err != nil {
		t.Fatalf("Save after fixing store: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "prompt_state.json")); err != nil {
		t.Fatalf("deck state not committed after successful save: %v", err)
	}
}

func TestDrawPromptResetsDeckWhenPromptsChange(t *testing.T) {
	s, dir := newTestSession(t)
	stale := core.Deck{Prompts: []string{"old-a", "old-b"}, Remaining: []int{1}, Used: []int{0}}
	if err := s.Store.SaveDeckState(stale); err != nil {
		t.Fatal(err)
	}
	if err := s.OpenDate("2026-06-11"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prompts.txt"), []byte("new-only\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := s.DrawPrompt("ignored"); err != nil {
		t.Fatalf("DrawPrompt: %v", err)
	}
	if s.Entry.Prompt != "new-only" {
		t.Fatalf("prompt = %q, want new-only", s.Entry.Prompt)
	}
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	deck, err := s.Store.LoadDeckState()
	if err != nil {
		t.Fatal(err)
	}
	if len(deck.Prompts) != 1 || deck.Prompts[0] != "new-only" {
		t.Fatalf("deck prompts = %#v, want reset to new prompts", deck.Prompts)
	}
}
