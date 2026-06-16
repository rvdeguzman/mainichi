package ui

import "github.com/charmbracelet/lipgloss"

var screenTitleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241")).
	Align(lipgloss.Center)

var screenHelpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241")).
	Align(lipgloss.Center)

func promptSourceLabel(source string) string {
	switch source {
	case "ai":
		return "ai (deprecated)"
	default:
		return source
	}
}

func promptSourceDescription(source string) string {
	switch source {
	case "stoic":
		return "use the Daily Stoic heading"
	case "deck":
		return "draw from the prompt deck"
	case "ai":
		return "generate with AI (deprecated)"
	default:
		return source
	}
}
