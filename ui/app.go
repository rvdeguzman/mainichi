package ui

import (
	"fmt"
	"mainichi/adapters"
	"mainichi/app"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	ViewWriter = iota
	ViewCalendar
	ViewRecent
	ViewConfig
)

type switchViewMsg struct {
	view int
	date string
}

type configDoneMsg struct {
	result ConfigResult
}

type aiPromptMsg struct {
	prompt string
	err    error
}

func fetchAIPrompt(apiKey string) tea.Cmd {
	return func() tea.Msg {
		prompt, err := adapters.GeneratePrompt(apiKey)
		return aiPromptMsg{prompt: prompt, err: err}
	}
}

type AppModel struct {
	store          adapters.Store
	session        *app.Session
	view           int
	writer         WriterModel
	calendar       CalendarModel
	recent         RecentModel
	config         ConfigModel
	width          int
	height         int
	aiPromptAPIKey string
}

func NewAppModel(store adapters.Store, session *app.Session, initialView int, aiPromptAPIKey string) AppModel {
	m := AppModel{
		store:          store,
		session:        session,
		view:           initialView,
		aiPromptAPIKey: aiPromptAPIKey,
	}
	switch initialView {
	case ViewWriter:
		m.writer = NewWriterModel(session)
		if aiPromptAPIKey != "" {
			m.writer.loadingPrompt = true
		}
	case ViewCalendar:
		m.calendar = NewCalendarModel(session)
	case ViewRecent:
		m.recent = NewRecentModel(session)
	case ViewConfig:
		m.config = NewConfigModel(session.Config)
	}
	return m
}

func (m AppModel) Init() tea.Cmd {
	switch m.view {
	case ViewWriter:
		cmds := []tea.Cmd{m.writer.Init()}
		if m.aiPromptAPIKey != "" {
			cmds = append(cmds, fetchAIPrompt(m.aiPromptAPIKey))
		}
		return tea.Batch(cmds...)
	case ViewCalendar:
		return m.calendar.Init()
	case ViewRecent:
		return m.recent.Init()
	case ViewConfig:
		return m.config.Init()
	}
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m.updateChild(msg)

	case switchViewMsg:
		if msg.date != "" {
			m.session.OpenDate(msg.date)
		}
		m.view = msg.view
		switch msg.view {
		case ViewWriter:
			m.writer = NewWriterModel(m.session)
		case ViewCalendar:
			m.calendar = NewCalendarModel(m.session)
		case ViewRecent:
			m.recent = NewRecentModel(m.session)
		case ViewConfig:
			m.config = NewConfigModel(m.session.Config)
		}
		return m, m.childInit()

	case configDoneMsg:
		m.applyConfig(msg.result)
		m.view = ViewWriter
		m.writer = NewWriterModel(m.session)
		return m, m.childInit()

	case aiPromptMsg:
		if msg.err != nil {
			fmt.Fprintf(os.Stderr, "warning: AI prompt failed: %v\n", msg.err)
		} else {
			m.session.Entry.Prompt = msg.prompt
			m.session.Save()
		}
		m.writer.loadingPrompt = false
		return m, nil

	default:
		return m.updateChild(msg)
	}
}

func (m AppModel) updateChild(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.view {
	case ViewWriter:
		var child tea.Model
		child, cmd = m.writer.Update(msg)
		m.writer = child.(WriterModel)
	case ViewCalendar:
		var child tea.Model
		child, cmd = m.calendar.Update(msg)
		m.calendar = child.(CalendarModel)
	case ViewRecent:
		var child tea.Model
		child, cmd = m.recent.Update(msg)
		m.recent = child.(RecentModel)
	case ViewConfig:
		var child tea.Model
		child, cmd = m.config.Update(msg)
		m.config = child.(ConfigModel)
	}
	return m, cmd
}

func (m AppModel) childInit() tea.Cmd {
	var cmds []tea.Cmd
	// Re-send the current window size to the new child
	if m.width > 0 || m.height > 0 {
		cmds = append(cmds, func() tea.Msg {
			return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		})
	}
	switch m.view {
	case ViewWriter:
		cmds = append(cmds, m.writer.Init())
	case ViewCalendar:
		cmds = append(cmds, m.calendar.Init())
	case ViewRecent:
		cmds = append(cmds, m.recent.Init())
	case ViewConfig:
		cmds = append(cmds, m.config.Init())
	}
	return tea.Batch(cmds...)
}

func (m *AppModel) applyConfig(res ConfigResult) {
	changed := false
	if res.Minimum != nil {
		m.session.Config.Minimum = *res.Minimum
		changed = true
	}
	if res.AutoPrompt != nil {
		m.session.Config.AutoPrompt = *res.AutoPrompt
		changed = true
	}
	if res.PromptSource != nil {
		m.session.Config.PromptSource = *res.PromptSource
		changed = true
	}
	if res.ResetDeck {
		if err := m.store.DeleteDeckState(); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: could not reset deck: %v\n", err)
		}
	}
	if changed {
		if err := m.store.SaveConfig(m.session.Config); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save config: %v\n", err)
		}
	}
}

func (m AppModel) View() string {
	switch m.view {
	case ViewWriter:
		return m.writer.View()
	case ViewCalendar:
		return m.calendar.View()
	case ViewRecent:
		return m.recent.View()
	case ViewConfig:
		return m.config.View()
	}
	return ""
}
