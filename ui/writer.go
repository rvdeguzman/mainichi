package ui

import (
	"fmt"
	"mainichi/app"
	"mainichi/core"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const cardWidth = 70

const (
	modeNormal = iota
	modePalette
	modeHelp
)

type paletteCommand struct {
	name   string
	desc   string
	action string
}

var commands = []paletteCommand{
	{"config", "settings", "config"},
	{"date", "calendar view", "date"},
	{"recent", "recent entries", "recent"},
	{"help", "show keybindings", "help"},
	{"exit", "save and quit", "quit"},
}

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

	palBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Width(40).
			Padding(0, 1)

	palPromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	palActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)

	palItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	palDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 2)

	helpTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)

	helpTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type promptTickMsg time.Time

type WriterModel struct {
	session          *app.Session
	textarea         textarea.Model
	width            int
	height           int
	mode             int
	paletteInput     string
	paletteCursor    int
	paletteFiltered  []paletteCommand
	paletteSearching bool
	loadingPrompt    bool
	spinnerFrame     int
}

func NewWriterModel(session *app.Session) WriterModel {
	ta := textarea.New()
	ta.Placeholder = "..."
	ta.SetValue(session.Entry.Body)
	ta.ShowLineNumbers = false
	ta.SetWidth(cardWidth - 4)
	ta.SetHeight(8)
	ta.CharLimit = 0
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("234"))
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.Focus()

	return WriterModel{
		session:         session,
		textarea:        ta,
		paletteFiltered: commands,
	}
}

func promptTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return promptTickMsg(t)
	})
}

func (m WriterModel) Init() tea.Cmd {
	cmds := []tea.Cmd{textarea.Blink}
	if m.loadingPrompt {
		cmds = append(cmds, promptTick())
	}
	return tea.Batch(cmds...)
}

func (m WriterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case promptTickMsg:
		if m.loadingPrompt {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
			return m, promptTick()
		}
		return m, nil
	}

	switch m.mode {
	case modePalette:
		return m.updatePalette(msg)
	case modeHelp:
		return m.updateHelp(msg)
	default:
		return m.updateNormal(msg)
	}
}

func (m WriterModel) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		case "esc":
			m.mode = modePalette
			m.paletteInput = ""
			m.paletteCursor = 0
			m.paletteFiltered = commands
			m.textarea.Blur()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m WriterModel) updatePalette(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.session.Entry.Body = m.textarea.Value()
			m.session.Save()
			return m, tea.Quit
		case "esc":
			if m.paletteSearching {
				m.paletteSearching = false
				m.paletteInput = ""
				m.paletteCursor = 0
				m.paletteFiltered = commands
				return m, nil
			}
			m.mode = modeNormal
			m.textarea.Focus()
			return m, nil
		case "enter":
			if len(m.paletteFiltered) > 0 {
				selected := m.paletteFiltered[m.paletteCursor]
				switch selected.action {
				case "help":
					m.mode = modeHelp
					return m, nil
				case "quit":
					m.session.Entry.Body = m.textarea.Value()
					m.session.Save()
					return m, tea.Quit
				case "config":
					m.session.Entry.Body = m.textarea.Value()
					m.session.Save()
					return m, func() tea.Msg { return switchViewMsg{view: ViewConfig} }
				case "date":
					m.session.Entry.Body = m.textarea.Value()
					m.session.Save()
					return m, func() tea.Msg { return switchViewMsg{view: ViewCalendar} }
				case "recent":
					m.session.Entry.Body = m.textarea.Value()
					m.session.Save()
					return m, func() tea.Msg { return switchViewMsg{view: ViewRecent} }
				}
			}
			return m, nil
		case "up":
			if m.paletteCursor > 0 {
				m.paletteCursor--
			}
			return m, nil
		case "down":
			if m.paletteCursor < len(m.paletteFiltered)-1 {
				m.paletteCursor++
			}
			return m, nil
		case "k":
			if !m.paletteSearching {
				if m.paletteCursor > 0 {
					m.paletteCursor--
				}
				return m, nil
			}
			fallthrough
		case "j":
			if !m.paletteSearching {
				if m.paletteCursor < len(m.paletteFiltered)-1 {
					m.paletteCursor++
				}
				return m, nil
			}
			fallthrough
		case "/":
			if !m.paletteSearching {
				m.paletteSearching = true
				return m, nil
			}
			// fall through to typing below
			fallthrough
		default:
			if !m.paletteSearching {
				return m, nil
			}
			if msg.String() == "backspace" {
				if len(m.paletteInput) > 0 {
					m.paletteInput = m.paletteInput[:len(m.paletteInput)-1]
					m.filterPalette()
				}
				return m, nil
			}
			for _, r := range msg.String() {
				if unicode.IsPrint(r) {
					m.paletteInput += string(r)
				}
			}
			m.filterPalette()
			return m, nil
		}
	}
	return m, nil
}

func (m *WriterModel) filterPalette() {
	input := strings.ToLower(m.paletteInput)
	var filtered []paletteCommand
	for _, cmd := range commands {
		if strings.Contains(strings.ToLower(cmd.name), input) {
			filtered = append(filtered, cmd)
		}
	}
	m.paletteFiltered = filtered
	if m.paletteCursor >= len(m.paletteFiltered) {
		m.paletteCursor = max(0, len(m.paletteFiltered)-1)
	}
}

func (m WriterModel) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.mode = modeNormal
			m.textarea.Focus()
			return m, nil
		}
	}
	return m, nil
}

func (m WriterModel) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.mode {
	case modePalette:
		return m.viewPalette()
	case modeHelp:
		return m.viewHelp()
	default:
		return m.viewWriter()
	}
}

func (m WriterModel) viewWriter() string {
	// Title
	title := titleStyle.Width(cardWidth).Render(
		fmt.Sprintf("mainichi — %s", m.session.Entry.Date),
	)

	// Prompt (optional)
	var promptLine string
	if m.loadingPrompt && m.session.Entry.Prompt == "" {
		frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
		promptLine = promptStyle.Width(cardWidth).Render(frame + " generating prompt…")
	} else if m.session.Entry.Prompt != "" {
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

func (m WriterModel) viewPalette() string {
	var lines []string

	if m.paletteSearching {
		cursor := "█"
		inputLine := palPromptStyle.Render("> " + m.paletteInput + cursor)
		lines = append(lines, inputLine)
	}

	// Filtered commands
	for i, cmd := range m.paletteFiltered {
		name := cmd.name
		desc := palDescStyle.Render("  " + cmd.desc)
		if i == m.paletteCursor {
			line := palActiveStyle.Render("  ▸ "+name) + desc
			lines = append(lines, line)
		} else {
			line := palItemStyle.Render("    "+name) + desc
			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n")
	box := palBoxStyle.Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m WriterModel) viewHelp() string {
	var b strings.Builder

	b.WriteString(helpTitleStyle.Render("keybindings"))
	b.WriteString("\n\n")
	b.WriteString(helpTextStyle.Render("  ctrl+s      save"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("  ctrl+c      save and quit"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("  esc         command palette"))
	b.WriteString("\n\n")
	b.WriteString(helpTextStyle.Render("  palette:"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("    type        filter commands"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("    ↑/↓         navigate"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("    enter       select"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("    esc         dismiss"))
	b.WriteString("\n\n")
	b.WriteString(helpTextStyle.Render("  commands:"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("    config      settings"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("    date        calendar view"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("    recent      recent entries"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("    help        this help"))
	b.WriteString("\n")
	b.WriteString(helpTextStyle.Render("    exit        save and quit"))

	box := helpBoxStyle.Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
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
