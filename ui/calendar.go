package ui

import (
	"fmt"
	"mainichi/app"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	calTitleStyle = titleStyle

	calHeaderStyle = sectionStyle

	calDayStyle = itemStyle.
			Width(3).
			Align(lipgloss.Center)

	calSelectedStyle = activeStyle.
				Width(3).
				Align(lipgloss.Center)

	calHelpStyle = helpTextStyle.
			Align(lipgloss.Center)
)

type CalendarModel struct {
	session *app.Session
	year    int
	month   time.Month
	cursor  int // day of month (1-based)
	width   int
	height  int
}

func NewCalendarModel(session *app.Session) CalendarModel {
	now := time.Now()
	return CalendarModel{
		session: session,
		year:    now.Year(),
		month:   now.Month(),
		cursor:  now.Day(),
	}
}

func (m CalendarModel) Init() tea.Cmd {
	return nil
}

func (m CalendarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc":
			return m, func() tea.Msg { return switchViewMsg{view: ViewWriter} }
		case "left", "h":
			m.cursor--
			if m.cursor < 1 {
				m.prevMonth()
				m.cursor = daysInMonth(m.year, m.month)
			}
		case "right", "l":
			m.cursor++
			if m.cursor > daysInMonth(m.year, m.month) {
				m.nextMonth()
				m.cursor = 1
			}
		case "up", "k":
			m.cursor -= 7
			if m.cursor < 1 {
				m.prevMonth()
				m.cursor += daysInMonth(m.year, m.month)
				if m.cursor < 1 {
					m.cursor = 1
				}
			}
		case "down", "j":
			m.cursor += 7
			max := daysInMonth(m.year, m.month)
			if m.cursor > max {
				m.cursor -= max
				m.nextMonth()
				if m.cursor > daysInMonth(m.year, m.month) {
					m.cursor = daysInMonth(m.year, m.month)
				}
			}
		case "[":
			m.prevMonth()
			max := daysInMonth(m.year, m.month)
			if m.cursor > max {
				m.cursor = max
			}
		case "]":
			m.nextMonth()
			max := daysInMonth(m.year, m.month)
			if m.cursor > max {
				m.cursor = max
			}
		case "enter":
			date := fmt.Sprintf("%04d-%02d-%02d", m.year, int(m.month), m.cursor)
			return m, func() tea.Msg { return switchViewMsg{view: ViewWriter, date: date} }
		}
	}

	return m, nil
}

func (m *CalendarModel) prevMonth() {
	m.month--
	if m.month < time.January {
		m.month = time.December
		m.year--
	}
}

func (m *CalendarModel) nextMonth() {
	m.month++
	if m.month > time.December {
		m.month = time.January
		m.year++
	}
}

func (m CalendarModel) View() string {
	if m.width == 0 {
		return ""
	}

	entries := m.session.ListEntries(m.year, m.month)

	// Title
	title := calTitleStyle.Width(cardWidth).Render(
		fmt.Sprintf("%s %d", m.month.String(), m.year),
	)

	// Day headers
	header := calHeaderStyle.Render("Mo Tu We Th Fr Sa Su")

	// Build grid
	first := time.Date(m.year, m.month, 1, 0, 0, 0, 0, time.Local)
	offset := int(first.Weekday())
	if offset == 0 {
		offset = 7
	}
	offset-- // Monday = 0

	days := daysInMonth(m.year, m.month)
	var rows []string
	var row []string

	// Leading blanks
	for i := 0; i < offset; i++ {
		row = append(row, calDayStyle.Render(" "))
	}

	for day := 1; day <= days; day++ {
		marker := markerFor(day, entries)
		cell := marker
		if day == m.cursor {
			cell = calSelectedStyle.Render(marker)
		} else {
			cell = calDayStyle.Render(marker)
		}
		row = append(row, cell)

		if (offset+day)%7 == 0 {
			rows = append(rows, strings.Join(row, ""))
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, strings.Join(row, ""))
	}

	// wip disabling help for now
	// help := calHelpStyle.Width(21).Render("hjkl navigate\n[ ] month\nenter open\nq quit")

	sections := []string{title, "", header}
	sections = append(sections, rows...)
	sections = append(sections, "", calHelpStyle.Render("←/→ month • esc back • ctrl+c quit"))

	block := lipgloss.JoinVertical(lipgloss.Center, sections...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
}

func markerFor(day int, entries map[int]app.CalendarEntry) string {
	e, ok := entries[day]
	if !ok {
		return "·"
	}
	if e.WordCount >= e.Minimum {
		return "●"
	}
	return "◐"
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}
