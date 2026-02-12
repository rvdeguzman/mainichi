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
		runWriter(store, session)

	case cmd == "prompt":
		// Open today with a deck prompt
		session.OpenToday()
		if err := session.DrawPrompt(mainichi.DefaultPrompts); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not draw prompt: %v\n", err)
		}
		runWriter(store, session)

	case cmd == "config":
		runConfig(store, cfg)

	case cmd == "date":
		date := runCalendar(session)
		if date != "" {
			session.OpenDate(date)
			runWriter(store, session)
		}

	case cmd == "recent":
		date := runRecent(session)
		if date != "" {
			session.OpenDate(date)
			runWriter(store, session)
		}

	case dateRe.MatchString(cmd):
		// Open specific date
		session.OpenDate(cmd)
		runWriter(store, session)

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

func runWriter(store adapters.Store, session *app.Session) {
	for {
		model := ui.NewWriterModel(session)
		p := tea.NewProgram(model, tea.WithAltScreen())
		m, err := p.Run()
		if err != nil {
			log.Fatal(err)
		}

		wm, ok := m.(ui.WriterModel)
		if !ok {
			return
		}

		switch wm.Action() {
		case "config":
			runConfig(store, session.Config)
			// Reload config in case it changed
			cfg, err := store.LoadConfig()
			if err == nil {
				session.Config = cfg
			}

		case "date":
			date := runCalendar(session)
			if date != "" {
				session.OpenDate(date)
			}

		case "recent":
			date := runRecent(session)
			if date != "" {
				session.OpenDate(date)
			}

		default:
			// "" or "quit" — exit
			return
		}
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

func runCalendar(session *app.Session) string {
	cal := ui.NewCalendarModel(session)
	p := tea.NewProgram(cal, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	if cm, ok := m.(ui.CalendarModel); ok {
		return cm.SelectedDate()
	}
	return ""
}

func runRecent(session *app.Session) string {
	recent := ui.NewRecentModel(session)
	p := tea.NewProgram(recent, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	if rm, ok := m.(ui.RecentModel); ok {
		return rm.SelectedDate()
	}
	return ""
}
