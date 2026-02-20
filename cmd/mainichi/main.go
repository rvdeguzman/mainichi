package main

import (
	"fmt"
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
	var aiPromptAPIKey string

	switch {
	case cmd == "":
		session.OpenToday()
		if cfg.AutoPrompt {
			switch cfg.PromptSource {
			case "ai":
				apiKey := os.Getenv("OPENAI_API_KEY")
				if apiKey != "" {
					aiPromptAPIKey = apiKey
				} else {
					if err := session.DrawPrompt(mainichi.DefaultPrompts); err != nil {
						fmt.Fprintf(os.Stderr, "warning: could not draw prompt: %v\n", err)
					}
				}
			case "stoic":
				session.SetStoicPrompt(mainichi.StoicHeadings)
			default:
				if err := session.DrawPrompt(mainichi.DefaultPrompts); err != nil {
					fmt.Fprintf(os.Stderr, "warning: could not draw prompt: %v\n", err)
				}
			}
		}
		initialView = ui.ViewWriter

	case cmd == "prompt":
		session.OpenToday()
		if err := session.DrawPrompt(mainichi.DefaultPrompts); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not draw prompt: %v\n", err)
		}
		initialView = ui.ViewWriter

	case cmd == "ai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "error: OPENAI_API_KEY not set\n")
			os.Exit(1)
		}
		session.OpenToday()
		aiPromptAPIKey = apiKey
		initialView = ui.ViewWriter

	case cmd == "stoic":
		session.OpenToday()
		session.SetStoicPrompt(mainichi.StoicHeadings)
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
		fmt.Fprintf(os.Stderr, "Usage: mainichi [command]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  (none)        Open today's entry\n")
		fmt.Fprintf(os.Stderr, "  prompt        Open today with a writing prompt\n")
		fmt.Fprintf(os.Stderr, "  ai            Open today with an AI-generated prompt\n")
		fmt.Fprintf(os.Stderr, "  stoic         Open today with the Daily Stoic heading\n")
		fmt.Fprintf(os.Stderr, "  config        Configure word count minimum\n")
		fmt.Fprintf(os.Stderr, "  date          Open calendar view\n")
		fmt.Fprintf(os.Stderr, "  recent        Browse recent entries\n")
		fmt.Fprintf(os.Stderr, "  YYYY-MM-DD    Open a specific date's entry\n")
		os.Exit(1)
	}

	model := ui.NewAppModel(store, session, initialView, aiPromptAPIKey, mainichi.DefaultPrompts, mainichi.StoicHeadings)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
