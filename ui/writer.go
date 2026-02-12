package ui

import (
	"fmt"
	"mainichi/app"
	"mainichi/core"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const cardWidth = 70

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true).
			Align(lipgloss.Center)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1)

	barFilledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	barEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("236"))
)

type WriterModel struct {
	session  *app.Session
	textarea textarea.Model
	width    int
	height   int
}

func NewWriterModel(session *app.Session) WriterModel {
	ta := textarea.New()
	ta.Placeholder = "Begin writing..."
	ta.SetValue(session.Entry.Body)
	ta.ShowLineNumbers = false
	ta.SetWidth(cardWidth - 4)
	ta.SetHeight(8)
	ta.CharLimit = 0
	ta.Focus()

	return WriterModel{
		session:  session,
		textarea: ta,
	}
}

func (m WriterModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m WriterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			m.session.Entry.Body = m.textarea.Value()
			m.session.Save()
			return m, nil
		case "ctrl+c":
			m.session.Entry.Body = m.textarea.Value()
			m.session.Save()
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m WriterModel) View() string {
	if m.width == 0 {
		return ""
	}

	// Title
	title := titleStyle.Width(cardWidth).Render(
		fmt.Sprintf("mainichi — %s", m.session.Entry.Date),
	)

	// Prompt (optional)
	var promptLine string
	if m.session.Entry.Prompt != "" {
		promptLine = promptStyle.Width(cardWidth).Render(
			m.session.Entry.Prompt,
		)
	}

	// Card
	card := cardStyle.Width(cardWidth).Render(m.textarea.View())

	// Progress bar
	wc := core.WordCount(m.textarea.Value())
	bar := renderBar(wc, m.session.Config.Minimum, cardWidth-4)

	// Compose
	var sections []string
	sections = append(sections, title)
	if promptLine != "" {
		sections = append(sections, promptLine)
	}
	sections = append(sections, card)
	sections = append(sections, lipgloss.NewStyle().Align(lipgloss.Center).Width(cardWidth).Render(bar))

	block := lipgloss.JoinVertical(lipgloss.Center, sections...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
}

func renderBar(words, minimum, width int) string {
	if minimum <= 0 {
		minimum = 1
	}
	ratio := float64(words) / float64(minimum)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	empty := width - filled

	bar := barFilledStyle.Render(strings.Repeat("▓", filled)) +
		barEmptyStyle.Render(strings.Repeat("░", empty))
	return bar
}
