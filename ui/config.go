package ui

import (
	"fmt"
	"mainichi/core"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var presets = []int{150, 250, 500, 750}

var (
	cfgTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center)

	cfgLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Align(lipgloss.Center)

	cfgItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	cfgActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)

	cfgCurrentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	cfgHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center)
)

type ConfigModel struct {
	current  int // current config value
	cursor   int // 0..len(presets) where last index = custom
	editing  bool
	input    textinput.Model
	saved    bool
	selected int // the value chosen, 0 if none
	width    int
	height   int
}

func NewConfigModel(cfg core.Config) ConfigModel {
	ti := textinput.New()
	ti.Placeholder = "number"
	ti.CharLimit = 5
	ti.Width = 6

	// Find cursor position matching current value
	cursor := len(presets) // default to custom
	for i, p := range presets {
		if p == cfg.Minimum {
			cursor = i
			break
		}
	}

	return ConfigModel{
		current: cfg.Minimum,
		cursor:  cursor,
		input:   ti,
	}
}

func (m ConfigModel) Init() tea.Cmd {
	return nil
}

func (m ConfigModel) Selected() (int, bool) {
	return m.selected, m.saved
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.editing {
			return m.updateEditing(msg)
		}
		return m.updateNavigating(msg)
	}

	return m, nil
}

func (m ConfigModel) updateNavigating(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := len(presets) // presets + custom

	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < maxIdx {
			m.cursor++
		}
	case "enter":
		if m.cursor < len(presets) {
			m.selected = presets[m.cursor]
			m.saved = true
			return m, tea.Quit
		}
		// Enter custom editing mode
		m.editing = true
		m.input.SetValue("")
		m.input.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

func (m ConfigModel) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editing = false
		m.input.Blur()
		return m, nil
	case "enter":
		val, err := strconv.Atoi(m.input.Value())
		if err != nil || val <= 0 {
			return m, nil
		}
		m.selected = val
		m.saved = true
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	// Filter non-numeric input
	filtered := ""
	for _, r := range m.input.Value() {
		if r >= '0' && r <= '9' {
			filtered += string(r)
		}
	}
	if filtered != m.input.Value() {
		m.input.SetValue(filtered)
	}

	return m, cmd
}

func (m ConfigModel) View() string {
	if m.width == 0 {
		return ""
	}

	w := cardWidth

	title := cfgTitleStyle.Width(w).Render("mainichi — config")
	label := cfgLabelStyle.Width(w).Render("Word count minimum:")

	var items []string
	for i, p := range presets {
		line := fmt.Sprintf("  %d", p)
		if p == m.current {
			line += " (current)"
		}
		if i == m.cursor {
			line = fmt.Sprintf("▸ %d", p)
			if p == m.current {
				line += " (current)"
			}
			items = append(items, cfgActiveStyle.Render(line))
		} else if p == m.current {
			items = append(items, cfgCurrentStyle.Render(line))
		} else {
			items = append(items, cfgItemStyle.Render(line))
		}
	}

	// Custom option
	customIdx := len(presets)
	if m.cursor == customIdx {
		if m.editing {
			items = append(items, cfgActiveStyle.Render("▸ Custom: ")+m.input.View())
		} else {
			items = append(items, cfgActiveStyle.Render("▸ Custom: ___"))
		}
	} else {
		items = append(items, cfgItemStyle.Render("  Custom: ___"))
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)
	centeredMenu := lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(menu)

	help := cfgHelpStyle.Width(w).Render("↑↓ navigate  enter select  esc cancel")

	block := lipgloss.JoinVertical(lipgloss.Center,
		title, "", label, "", centeredMenu, "", help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
}
