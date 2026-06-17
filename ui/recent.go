package ui

import (
	"fmt"
	"mainichi/app"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	recentTitleStyle = titleStyle

	recentMutedStyle = itemStyle

	recentActiveStyle = activeStyle

	recentWordStyle = mutedStyle

	recentHelpStyle = helpTextStyle.
			Align(lipgloss.Center)

	recentPreviewBorder = boxStyle.Foreground(colorItem)
)

type RecentModel struct {
	session *app.Session
	entries []app.RecentEntry
	cursor  int
	width   int
	height  int
}

func NewRecentModel(session *app.Session) RecentModel {
	entries, _ := session.ListRecentEntries(30)
	return RecentModel{
		session: session,
		entries: entries,
	}
}

func (m RecentModel) Init() tea.Cmd {
	return nil
}

func (m RecentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.entries) > 0 {
				date := m.entries[m.cursor].Entry.Date
				return m, func() tea.Msg { return switchViewMsg{view: ViewWriter, date: date} }
			}
		}
	}

	return m, nil
}

func (m RecentModel) View() string {
	if m.width == 0 {
		return ""
	}

	title := recentTitleStyle.Width(cardWidth).Render("mainichi — recent")

	if len(m.entries) == 0 {
		empty := recentMutedStyle.Render("No entries yet.")
		block := lipgloss.JoinVertical(lipgloss.Center, title, "", empty)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
	}

	showPreview := m.width >= 60

	// Build list
	var listLines []string
	for i, re := range m.entries {
		minimum := re.Entry.Minimum
		if minimum <= 0 {
			minimum = m.session.Config.Minimum
		}
		marker := "●"
		if re.WordCount < minimum {
			marker = "◐"
		}

		line := fmt.Sprintf("%s  %s %d words", re.Entry.Date, marker, re.WordCount)

		if i == m.cursor {
			listLines = append(listLines, "  "+recentActiveStyle.Render("▸ "+line))
		} else {
			listLines = append(listLines, "    "+recentMutedStyle.Render(line))
		}
	}

	// Visible rows: leave room for title, blank, help, blank
	maxVisible := m.height - 6
	if maxVisible < 1 {
		maxVisible = 1
	}

	// Scroll window
	scrollTop := 0
	if m.cursor >= maxVisible {
		scrollTop = m.cursor - maxVisible + 1
	}
	scrollEnd := scrollTop + maxVisible
	if scrollEnd > len(listLines) {
		scrollEnd = len(listLines)
	}
	visibleList := listLines[scrollTop:scrollEnd]
	listBlock := strings.Join(visibleList, "\n")

	var content string
	if showPreview {
		// Preview pane
		previewWidth := m.width/2 - recentPreviewBorder.GetHorizontalFrameSize() - 2
		if previewWidth < 20 {
			previewWidth = 20
		}
		previewHeight := maxVisible - recentPreviewBorder.GetVerticalFrameSize()
		if previewHeight < 1 {
			previewHeight = 1
		}

		body := m.entries[m.cursor].Entry.Body
		preview := wrapAndTruncate(body, previewWidth, previewHeight)

		previewBox := recentPreviewBorder.
			Width(previewWidth).
			Height(previewHeight).
			Render(preview)

		content = lipgloss.JoinHorizontal(lipgloss.Top, listBlock, "  ", previewBox)
	} else {
		content = listBlock
	}

	// help := recentHelpStyle.Render("↑↓ navigate  enter open  q quit")

	block := lipgloss.JoinVertical(lipgloss.Center, title, "", content, "", recentHelpStyle.Render("↑/↓ scroll • esc back • ctrl+c quit"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
}

func wrapAndTruncate(text string, width, maxLines int) string {
	if width <= 0 || maxLines <= 0 {
		return ""
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	current := words[0]

	for _, w := range words[1:] {
		if len(current)+1+len(w) > width {
			lines = append(lines, current)
			current = w
			if len(lines) >= maxLines {
				break
			}
		} else {
			current += " " + w
		}
	}
	if len(lines) < maxLines {
		lines = append(lines, current)
	}

	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	return strings.Join(lines, "\n")
}
