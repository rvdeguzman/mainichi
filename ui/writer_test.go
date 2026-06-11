package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mainichi/adapters"
	"mainichi/app"
	"mainichi/core"

	tea "github.com/charmbracelet/bubbletea"
)

func failingSaveSession(t *testing.T) (*app.Session, string) {
	t.Helper()
	dir := t.TempDir()
	s := app.NewSession(adapters.NewStore(dir), core.DefaultConfig())
	if err := s.OpenDate("2026-06-11"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "entries"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	return s, dir
}

func TestWriterCtrlSSurfacesSaveError(t *testing.T) {
	s, _ := failingSaveSession(t)
	m := NewWriterModel(s)
	m.width = 80
	m.height = 24
	m.textarea.SetValue("draft")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd != nil {
		t.Fatalf("ctrl+s cmd = %v, want nil", cmd)
	}
	got := updated.(WriterModel)
	if got.status == "" || !strings.Contains(got.status, "couldn't save") {
		t.Fatalf("status = %q, want calm save error", got.status)
	}
	if !strings.Contains(got.View(), "couldn't save") {
		t.Fatalf("view did not surface save error: %q", got.View())
	}
}

func TestWriterQuitDoesNotQuitWhenSaveFails(t *testing.T) {
	s, _ := failingSaveSession(t)
	m := NewWriterModel(s)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		// expected: no quit command; keep user in writer
	} else if msg := cmd(); msg == tea.Quit() {
		t.Fatal("ctrl+c returned tea.Quit despite save failure")
	}
	got := updated.(WriterModel)
	if got.status == "" {
		t.Fatal("status empty, want save error")
	}
}

func TestPaletteNavigationDoesNotSwitchViewWhenSaveFails(t *testing.T) {
	s, _ := failingSaveSession(t)
	m := NewWriterModel(s)
	m.mode = modePalette
	m.paletteFiltered = []paletteCommand{{name: "config", action: "config"}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, ok := msg.(switchViewMsg); ok {
				t.Fatal("config command emitted switchViewMsg despite save failure")
			}
		}
	}
	got := updated.(WriterModel)
	if got.mode != modePalette {
		t.Fatalf("mode = %d, want palette to remain open", got.mode)
	}
	if got.status == "" {
		t.Fatal("status empty, want save error")
	}
}

func TestNewWriterModelSurfacesOpenError(t *testing.T) {
	s, _ := failingSaveSession(t)
	s.OpenError = os.ErrPermission
	m := NewWriterModel(s)
	m.width = 80
	m.height = 24
	if !strings.Contains(m.View(), "couldn't open entry") {
		t.Fatalf("view did not surface open error: %q", m.View())
	}
}
