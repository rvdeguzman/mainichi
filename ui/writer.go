package ui

import (
	"fmt"
	"mainichi/app"
	"mainichi/core"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const cardWidth = 70

const (
	modeNormal = iota
	modePalette
	modeHelp
	modePromptEdit
)

type paletteCommand struct {
	name   string
	desc   string
	action string
}

var mainCommands = []paletteCommand{
	{"config", "settings", "config"},
	{"date", "calendar view", "date"},
	{"recent", "recent entries", "recent"},
	{"prompt", "manage prompt", "prompt-submenu"},
	{"help", "show keybindings", "help"},
	{"exit", "save and quit", "quit"},
}

func promptSubCommands(hasPrompt bool) []paletteCommand {
	var cmds []paletteCommand
	if hasPrompt {
		cmds = append(cmds,
			paletteCommand{"edit", "modify current prompt", "prompt-edit"},
			paletteCommand{"delete", "remove prompt", "prompt-delete"},
		)
	}
	cmds = append(cmds,
		paletteCommand{"stoic", promptSourceDescription("stoic"), "prompt-stoic"},
		paletteCommand{"deck", promptSourceDescription("deck"), "prompt-deck"},
		paletteCommand{"ai", promptSourceDescription("ai"), "prompt-ai"},
	)
	return cmds
}

var (
	titleStyle = screenTitleStyle

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
	paletteSubmenu   bool // true when in prompt submenu
	loadingPrompt    bool
	spinnerFrame     int
	promptInput      textinput.Model
	status           string
}

func NewWriterModel(session *app.Session) WriterModel {
	ta := textarea.New()
	ta.Placeholder = "..."
	ta.SetValue(session.Entry.Body)
	ta.ShowLineNumbers = false
	ta.SetWidth(cardWidth - 4)
	ta.SetHeight(8)
	ta.CharLimit = 0
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.Focus()

	pi := textinput.New()
	pi.Placeholder = "enter prompt"
	pi.CharLimit = 200
	pi.Width = 36

	return WriterModel{
		session:         session,
		textarea:        ta,
		promptInput:     pi,
		paletteFiltered: mainCommands,
		status:          openStatus(session.OpenError),
	}
}

func openStatus(err error) string {
	if err == nil {
		return ""
	}
	return "couldn't open entry — saved copy left untouched"
}

func saveStatus(err error) string {
	if err == nil {
		return "saved"
	}
	return "couldn't save — please check your mainichi files"
}

func (m *WriterModel) setStatus(status string) {
	m.status = status
}

func (m *WriterModel) save() error {
	m.session.Entry.Body = m.textarea.Value()
	if err := m.session.Save(); err != nil {
		m.setStatus(saveStatus(err))
		return err
	}
	m.setStatus(saveStatus(nil))
	return nil
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
	case modePromptEdit:
		return m.updatePromptEdit(msg)
	default:
		return m.updateNormal(msg)
	}
}

func (m WriterModel) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			_ = m.save()
			return m, nil
		case "ctrl+c":
			if err := m.save(); err != nil {
				return m, nil
			}
			return m, tea.Quit
		case "esc":
			m.mode = modePalette
			m.paletteInput = ""
			m.paletteCursor = 0
			m.paletteSubmenu = false
			m.paletteFiltered = mainCommands
			m.paletteSearching = false
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
			if err := m.save(); err != nil {
				return m, nil
			}
			return m, tea.Quit
		case "esc":
			if m.paletteSearching {
				m.paletteSearching = false
				m.paletteInput = ""
				m.paletteCursor = 0
				if m.paletteSubmenu {
					m.paletteFiltered = promptSubCommands(m.session.Entry.Prompt != "")
				} else {
					m.paletteFiltered = mainCommands
				}
				return m, nil
			}
			if m.paletteSubmenu {
				// Back to main menu
				m.paletteSubmenu = false
				m.paletteCursor = 0
				m.paletteFiltered = mainCommands
				return m, nil
			}
			m.mode = modeNormal
			m.textarea.Focus()
			return m, nil
		case "enter":
			if len(m.paletteFiltered) > 0 {
				selected := m.paletteFiltered[m.paletteCursor]
				switch selected.action {
				case "prompt-submenu":
					m.paletteSubmenu = true
					m.paletteCursor = 0
					m.paletteInput = ""
					m.paletteSearching = false
					m.paletteFiltered = promptSubCommands(m.session.Entry.Prompt != "")
					return m, nil
				case "help":
					m.mode = modeHelp
					return m, nil
				case "quit":
					if err := m.save(); err != nil {
						return m, nil
					}
					return m, tea.Quit
				case "config":
					if err := m.save(); err != nil {
						return m, nil
					}
					return m, func() tea.Msg { return switchViewMsg{view: ViewConfig} }
				case "date":
					if err := m.save(); err != nil {
						return m, nil
					}
					return m, func() tea.Msg { return switchViewMsg{view: ViewCalendar} }
				case "recent":
					if err := m.save(); err != nil {
						return m, nil
					}
					return m, func() tea.Msg { return switchViewMsg{view: ViewRecent} }
				case "prompt-delete":
					m.session.Entry.Prompt = ""
					if err := m.save(); err != nil {
						return m, nil
					}
					m.mode = modeNormal
					m.textarea.Focus()
					return m, nil
				case "prompt-edit":
					m.mode = modePromptEdit
					m.promptInput.SetValue(m.session.Entry.Prompt)
					m.promptInput.Focus()
					return m, textinput.Blink
				case "prompt-deck", "prompt-ai", "prompt-stoic":
					// Clear existing prompt so regeneration works
					m.session.Entry.Prompt = ""
					action := strings.TrimPrefix(selected.action, "prompt-")
					if action == "ai" {
						m.loadingPrompt = true
					}
					m.mode = modeNormal
					m.textarea.Focus()
					return m, func() tea.Msg { return promptActionMsg{action: action} }
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
	var source []paletteCommand
	if m.paletteSubmenu {
		source = promptSubCommands(m.session.Entry.Prompt != "")
	} else {
		source = mainCommands
	}
	var filtered []paletteCommand
	for _, cmd := range source {
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

func (m WriterModel) updatePromptEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = modeNormal
			m.promptInput.Blur()
			m.textarea.Focus()
			return m, nil
		case "enter":
			m.session.Entry.Prompt = m.promptInput.Value()
			if err := m.save(); err != nil {
				return m, nil
			}
			m.mode = modeNormal
			m.promptInput.Blur()
			m.textarea.Focus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.promptInput, cmd = m.promptInput.Update(msg)
	return m, cmd
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
	case modePromptEdit:
		return m.viewPromptEdit()
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
	minimum := m.session.Entry.Minimum
	if minimum <= 0 {
		minimum = m.session.Config.Minimum
	}
	bar := renderBar(wc, minimum, cardWidth-4)

	// Compose
	var sections []string
	sections = append(sections, title)
	if promptLine != "" {
		sections = append(sections, promptLine)
	}
	if m.status != "" {
		sections = append(sections, palDescStyle.Width(cardWidth).Align(lipgloss.Center).Render(m.status))
	}
	sections = append(sections, card)
	sections = append(sections, lipgloss.NewStyle().Align(lipgloss.Center).Width(cardWidth).Render(bar))

	block := lipgloss.JoinVertical(lipgloss.Center, sections...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
}

func (m WriterModel) viewPalette() string {
	var lines []string

	if m.paletteSubmenu {
		lines = append(lines, palDescStyle.Render("  prompt ›"))
	}

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

func (m WriterModel) viewPromptEdit() string {
	var lines []string
	lines = append(lines, palPromptStyle.Render("  edit prompt"))
	lines = append(lines, "")
	lines = append(lines, "  "+m.promptInput.View())
	lines = append(lines, "")
	lines = append(lines, palDescStyle.Render("  enter save  esc cancel"))

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
