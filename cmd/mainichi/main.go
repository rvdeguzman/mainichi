package main

import (
	"fmt"
	"log"
	"os"
	"regexp"

	mainichi "mainichi"
	"mainichi/adapters"
	"mainichi/app"
	"mainichi/core"
	"mainichi/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func main() {
	store, err := adapters.DefaultStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.EnsureDirs(); err != nil {
		log.Fatal(err)
	}

	cfg, err := store.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
	}

	session := app.NewSession(store, cfg)

	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch {
	case cmd == "":
		// Open today's entry
		session.OpenToday()
		if cfg.AutoPrompt {
			if err := session.DrawPrompt(mainichi.DefaultPrompts); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not draw prompt: %v\n", err)
			}
		}
		runWriter(session)

	case cmd == "prompt":
		// Open today with a deck prompt
		session.OpenToday()
		if err := session.DrawPrompt(mainichi.DefaultPrompts); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not draw prompt: %v\n", err)
		}
		runWriter(session)

	case cmd == "config":
		runConfig(store, cfg)

	case cmd == "date":
		// Open calendar view
		runCalendar(session)

	case cmd == "recent":
		// Open recent entries view
		runRecent(session)

	case dateRe.MatchString(cmd):
		// Open specific date
		session.OpenDate(cmd)
		runWriter(session)

	default:
		fmt.Fprintf(os.Stderr, "Usage: mainichi [command]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  (none)        Open today's entry\n")
		fmt.Fprintf(os.Stderr, "  prompt        Open today with a writing prompt\n")
		fmt.Fprintf(os.Stderr, "  config        Configure word count minimum\n")
		fmt.Fprintf(os.Stderr, "  date          Open calendar view\n")
		fmt.Fprintf(os.Stderr, "  recent        Browse recent entries\n")
		fmt.Fprintf(os.Stderr, "  YYYY-MM-DD    Open a specific date's entry\n")
		os.Exit(1)
	}
}

func runWriter(session *app.Session) {
	model := ui.NewWriterModel(session)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func runConfig(store adapters.Store, cfg core.Config) {
	model := ui.NewConfigModel(cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	cm, ok := m.(ui.ConfigModel)
	if !ok {
		return
	}

	res := cm.Result()
	changed := false

	if res.Minimum != nil {
		cfg.Minimum = *res.Minimum
		changed = true
	}
	if res.AutoPrompt != nil {
		cfg.AutoPrompt = *res.AutoPrompt
		changed = true
	}
	if res.ResetDeck {
		if err := store.DeleteDeckState(); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: could not reset deck: %v\n", err)
		}
	}
	if changed {
		if err := store.SaveConfig(cfg); err != nil {
			log.Fatal(err)
		}
	}
}

func runCalendar(session *app.Session) {
	cal := ui.NewCalendarModel(session)
	p := tea.NewProgram(cal, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	// If user selected a date, open it in the writer
	if cm, ok := m.(ui.CalendarModel); ok && cm.SelectedDate() != "" {
		session.OpenDate(cm.SelectedDate())
		runWriter(session)
	}
}

func runRecent(session *app.Session) {
	recent := ui.NewRecentModel(session)
	p := tea.NewProgram(recent, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	if rm, ok := m.(ui.RecentModel); ok && rm.SelectedDate() != "" {
		session.OpenDate(rm.SelectedDate())
		runWriter(session)
	}
}
