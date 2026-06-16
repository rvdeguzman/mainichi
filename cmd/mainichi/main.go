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
		cmd = normalizeCommand(os.Args[1])
	}

	var initialView int
	initialPromptSource := ""
	aiPromptAPIKey := os.Getenv("OPENAI_API_KEY")

	switch {
	case cmd == "":
		session.OpenToday()
		if cfg.AutoPrompt {
			initialPromptSource = cfg.PromptSource
			if initialPromptSource == "" {
				initialPromptSource = "stoic"
			}
			if initialPromptSource == "ai" && aiPromptAPIKey == "" {
				fmt.Fprintf(os.Stderr, "warning: OPENAI_API_KEY not set; using stoic prompt instead\n")
				initialPromptSource = "stoic"
			}
		}
		initialView = ui.ViewWriter

	case cmd == "prompt":
		session.OpenToday()
		initialPromptSource = cfg.PromptSource
		if initialPromptSource == "" {
			initialPromptSource = "stoic"
		}
		if initialPromptSource == "ai" && aiPromptAPIKey == "" {
			fmt.Fprintf(os.Stderr, "warning: OPENAI_API_KEY not set; using stoic prompt instead\n")
			initialPromptSource = "stoic"
		}
		initialView = ui.ViewWriter

	case cmd == "ai":
		if aiPromptAPIKey == "" {
			fmt.Fprintf(os.Stderr, "error: OPENAI_API_KEY not set\n")
			os.Exit(1)
		}
		session.OpenToday()
		initialPromptSource = "ai"
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

	model := ui.NewAppModel(store, session, initialView, initialPromptSource, aiPromptAPIKey, mainichi.DefaultPrompts, mainichi.StoicHeadings)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func normalizeCommand(cmd string) string {
	if cmd == "--ai" {
		return "ai"
	}
	return cmd
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage: mainichi [command]\n\n")
	fmt.Fprintf(w, "Commands:\n")
	fmt.Fprintf(w, "  (none)        Open today's entry\n")
	fmt.Fprintf(w, "  prompt        Open today with the configured prompt source (stoic by default)\n")
	fmt.Fprintf(w, "  ai, --ai      Open today with an AI-generated prompt (deprecated)\n")
	fmt.Fprintf(w, "  stoic         Open today with the Daily Stoic heading\n")
	fmt.Fprintf(w, "  config        Configure writing settings\n")
	fmt.Fprintf(w, "  date          Open calendar view\n")
	fmt.Fprintf(w, "  recent        Browse recent entries\n")
	fmt.Fprintf(w, "  YYYY-MM-DD    Open a specific date's entry\n")
}
