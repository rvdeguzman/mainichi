package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	mainichi "mainichi"
	"mainichi/adapters"
	"mainichi/app"
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

	var initialView int
	initialPromptSource := ""

	switch {
	case cmd == "":
		session.OpenToday()
		if cfg.AutoPrompt {
			initialPromptSource = supportedPromptSource(cfg.PromptSource)
		}
		initialView = ui.ViewWriter

	case cmd == "prompt":
		session.OpenToday()
		initialPromptSource = supportedPromptSource(cfg.PromptSource)
		initialView = ui.ViewWriter

	case cmd == "stoic":
		session.OpenToday()
		initialPromptSource = "stoic"
		initialView = ui.ViewWriter

	case cmd == "config":
		initialView = ui.ViewConfig

	case cmd == "date":
		session.OpenToday()
		initialView = ui.ViewCalendar

	case cmd == "recent":
		session.OpenToday()
		initialView = ui.ViewRecent

	case dateRe.MatchString(cmd):
		session.OpenDate(cmd)
		initialView = ui.ViewWriter

	default:
		printUsage(os.Stderr)
		os.Exit(1)
	}

	model := ui.NewAppModel(store, session, initialView, initialPromptSource, mainichi.DefaultPrompts, mainichi.StoicHeadings)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func supportedPromptSource(source string) string {
	if source == "deck" {
		return source
	}
	return "stoic"
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage: mainichi [command]\n\n")
	fmt.Fprintf(w, "Commands:\n")
	fmt.Fprintf(w, "  (none)        Open today's entry\n")
	fmt.Fprintf(w, "  prompt        Open today with the configured prompt source (stoic by default)\n")
	fmt.Fprintf(w, "  stoic         Open today with the Daily Stoic heading\n")
	fmt.Fprintf(w, "  config        Configure writing settings\n")
	fmt.Fprintf(w, "  date          Open calendar view\n")
	fmt.Fprintf(w, "  recent        Browse recent entries\n")
	fmt.Fprintf(w, "  YYYY-MM-DD    Open a specific date's entry\n")
}
