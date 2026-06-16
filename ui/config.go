package ui

import (
	"fmt"
	"mainichi/core"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var presets = []int{150, 250, 300, 500, 750}

var (
	cfgTitleStyle = screenTitleStyle

	cfgSectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	cfgItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	cfgActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)

	cfgCurrentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	cfgHelpStyle = screenHelpStyle
)

type itemKind int

const (
	kindMinimumPreset itemKind = iota
	kindMinimumCustom
	kindAutoPrompt
	kindPromptSource
	kindResetDeck
)

type configItem struct {
	kind        itemKind
	value       int // preset value for kindMinimumPreset; bool as 0/1 for kindAutoPrompt
	label       string
	stringValue string // for kindPromptSource
}

type ConfigResult struct {
	Minimum      *int
	AutoPrompt   *bool
	PromptSource *string
	ResetDeck    bool
}

type ConfigModel struct {
	items               []configItem
	cursor              int
	editing             bool
	input               textinput.Model
	result              ConfigResult
	currentMinimum      int
	currentAutoPrompt   bool
	currentPromptSource string
	width               int
	height              int
}

func NewConfigModel(cfg core.Config) ConfigModel {
	ti := textinput.New()
	ti.Placeholder = "number"
	ti.CharLimit = 5
	ti.Width = 6

	var items []configItem

	// Word minimum section: current value first, then presets, then custom
	currentInPresets := false
	items = append(items, configItem{kind: kindMinimumPreset, value: cfg.Minimum, label: fmt.Sprintf("%d", cfg.Minimum)})
	for _, p := range presets {
		if p == cfg.Minimum {
			currentInPresets = true
			continue
		}
		items = append(items, configItem{kind: kindMinimumPreset, value: p, label: fmt.Sprintf("%d", p)})
	}
	if !currentInPresets {
		// Current value wasn't a preset — it's already first, add all presets
		// (already done above since we skip nothing)
	}
	items = append(items, configItem{kind: kindMinimumCustom, label: "Custom: ___"})

	// Auto-prompt section
	if cfg.AutoPrompt {
		items = append(items, configItem{kind: kindAutoPrompt, value: 1, label: "on"})
		items = append(items, configItem{kind: kindAutoPrompt, value: 0, label: "off"})
	} else {
		items = append(items, configItem{kind: kindAutoPrompt, value: 0, label: "off"})
		items = append(items, configItem{kind: kindAutoPrompt, value: 1, label: "on"})
	}

	// Prompt source section
	promptSource := cfg.PromptSource
	if promptSource == "" {
		promptSource = "stoic"
	}
	allSources := []string{"stoic", "deck", "ai"}
	for _, s := range allSources {
		items = append(items, configItem{kind: kindPromptSource, stringValue: s, label: promptSourceLabel(s)})
	}

	// Reset deck section
	items = append(items, configItem{kind: kindResetDeck, label: "Reset deck"})

	return ConfigModel{
		items:               items,
		cursor:              0,
		input:               ti,
		currentMinimum:      cfg.Minimum,
		currentAutoPrompt:   cfg.AutoPrompt,
		currentPromptSource: promptSource,
	}
}

func (m ConfigModel) Init() tea.Cmd {
	return nil
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
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return m, func() tea.Msg { return switchViewMsg{view: ViewWriter} }
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case "enter":
		item := m.items[m.cursor]
		switch item.kind {
		case kindMinimumPreset:
			if item.value != m.currentMinimum {
				v := item.value
				m.result.Minimum = &v
			}
			res := m.result
			return m, func() tea.Msg { return configDoneMsg{result: res} }
		case kindMinimumCustom:
			m.editing = true
			m.input.SetValue("")
			m.input.Focus()
			return m, textinput.Blink
		case kindAutoPrompt:
			b := item.value == 1
			if b != m.currentAutoPrompt {
				m.result.AutoPrompt = &b
			}
			res := m.result
			return m, func() tea.Msg { return configDoneMsg{result: res} }
		case kindPromptSource:
			if item.stringValue != m.currentPromptSource {
				s := item.stringValue
				m.result.PromptSource = &s
			}
			res := m.result
			return m, func() tea.Msg { return configDoneMsg{result: res} }
		case kindResetDeck:
			m.result.ResetDeck = true
			res := m.result
			return m, func() tea.Msg { return configDoneMsg{result: res} }
		}
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
		if val != m.currentMinimum {
			m.result.Minimum = &val
		}
		res := m.result
		return m, func() tea.Msg { return configDoneMsg{result: res} }
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

	var lines []string

	// Section header: Word minimum
	lines = append(lines, cfgSectionStyle.Render("  Word minimum"))

	for i, item := range m.items {
		// Insert section headers at transitions
		if item.kind == kindAutoPrompt && (i == 0 || m.items[i-1].kind != kindAutoPrompt) {
			lines = append(lines, "") // blank line separator
			lines = append(lines, cfgSectionStyle.Render("  Auto-prompt"))
		}
		if item.kind == kindPromptSource && (i == 0 || m.items[i-1].kind != kindPromptSource) {
			lines = append(lines, "") // blank line separator
			lines = append(lines, cfgSectionStyle.Render("  Prompt source"))
		}
		if item.kind == kindResetDeck && (i == 0 || m.items[i-1].kind != kindResetDeck) {
			lines = append(lines, "") // blank line separator
			lines = append(lines, cfgSectionStyle.Render("  Prompt deck"))
		}

		active := i == m.cursor
		line := m.renderItem(item, active)
		lines = append(lines, line)
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, lines...)
	centeredMenu := lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(menu)

	help := screenHelpStyle.Width(w).Render("←/→ change • enter save • esc back • ctrl+c quit")

	block := lipgloss.JoinVertical(lipgloss.Center,
		title, "", centeredMenu, "", help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
}

func (m ConfigModel) renderItem(item configItem, active bool) string {
	switch item.kind {
	case kindMinimumPreset:
		current := item.value == m.currentMinimum
		return m.renderChoice(fmt.Sprintf("%d", item.value), current, active)

	case kindMinimumCustom:
		if active {
			if m.editing {
				return cfgActiveStyle.Render("  ▸ Custom: ") + m.input.View()
			}
			return cfgActiveStyle.Render("  ▸ Custom: ___")
		}
		return cfgItemStyle.Render("    Custom: ___")

	case kindAutoPrompt:
		current := (item.value == 1) == m.currentAutoPrompt
		return m.renderChoice(item.label, current, active)

	case kindPromptSource:
		current := item.stringValue == m.currentPromptSource
		return m.renderChoice(item.label, current, active)

	case kindResetDeck:
		if active {
			return cfgActiveStyle.Render("  ▸ Reset deck")
		}
		return cfgItemStyle.Render("    Reset deck")
	}
	return ""
}

func (m ConfigModel) renderChoice(label string, current, active bool) string {
	suffix := ""
	if current {
		suffix = " (current)"
	}

	if active {
		return cfgActiveStyle.Render(fmt.Sprintf("  ▸ %s%s", label, suffix))
	}
	if current {
		return cfgCurrentStyle.Render(fmt.Sprintf("    %s%s", label, suffix))
	}
	return cfgItemStyle.Render(fmt.Sprintf("    %s", label))
}
