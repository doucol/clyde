package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type helpModel struct {
	width  int
	height int
}

func newHelpModel() helpModel {
	return helpModel{}
}

func (m helpModel) Init() tea.Cmd { return nil }

func (m helpModel) setSize(w, h int) helpModel {
	m.width = w
	m.height = h
	return m
}

func (m helpModel) Update(msg tea.Msg) (helpModel, bool) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "esc", "?":
			return m, true
		}
	}
	return m, false
}

func (m helpModel) View() string {
	title := styleOverlayTitle.Render("Help — Key Commands")
	maxKeyLen := 0
	for _, e := range helpEntries {
		if len(e.Key) > maxKeyLen {
			maxKeyLen = len(e.Key)
		}
	}
	lines := []string{title, ""}
	for _, e := range helpEntries {
		line := lipgloss.JoinHorizontal(lipgloss.Top,
			styleMenuKey.Render(padRight(e.Key, maxKeyLen+2)),
			styleStatusVal.Render(e.Description),
		)
		lines = append(lines, line)
	}
	lines = append(lines, "", styleHelp.Render("esc or ? to close"))
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	boxed := styleBorder.Padding(1, 2).Render(body)
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, boxed)
}
