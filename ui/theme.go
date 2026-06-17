package ui

import "github.com/charmbracelet/lipgloss"

// Layout constants shared across every view. Centralizing them removes the
// width/padding drift that crept in between the writer, calendar, recent and
// config screens.
const (
	cardWidth    = 70 // primary content width: writer card, config menu, page titles
	paletteWidth = 40 // command palette and modal boxes
)

// Centralized colour ramp (ANSI 256). One ramp, referenced everywhere, so the
// views stay visually coherent instead of each picking its own greys.
const (
	colorBorder = lipgloss.Color("238") // rounded borders + least-prominent text
	colorTitle  = lipgloss.Color("241") // muted, non-bold page titles
	colorMuted  = lipgloss.Color("243") // section labels + secondary text
	colorItem   = lipgloss.Color("245") // readable unselected list/menu items
	colorActive = lipgloss.Color("255") // selected / active emphasis
	colorPrompt = lipgloss.Color("240") // textarea prompt cursor
)

var (
	// titleStyle is the single muted, non-bold, centered page-title style.
	// Every view's title uses this so they read identically.
	titleStyle = lipgloss.NewStyle().
			Foreground(colorTitle).
			Align(lipgloss.Center)

	// itemStyle is the single readable style for unselected list/menu items.
	itemStyle = lipgloss.NewStyle().
			Foreground(colorItem)

	// activeStyle is the single style for the selected/active item.
	activeStyle = lipgloss.NewStyle().
			Foreground(colorActive).
			Bold(true)

	// mutedStyle is for section headers and secondary text.
	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Shared aliases for the remaining repeated text roles.
	helpTitleStyle = titleStyle
	helpTextStyle  = mutedStyle
	sectionStyle   = mutedStyle

	// faintStyle is for the least-prominent text (descriptions, status lines).
	faintStyle = lipgloss.NewStyle().
			Foreground(colorBorder)

	// borderStyle is the shared rounded border with a consistent colour. All
	// bordered surfaces derive from this so borders never drift apart.
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder)

	// boxStyle is a calm, padded, rounded box for cards and modal/palette panes.
	boxStyle = borderStyle.Padding(0, 1)
)
