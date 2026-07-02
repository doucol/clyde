package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/doucol/clyde/internal/kube"
)

type homeModel struct {
	kc       *kube.KubeconfigInfo
	loadErr  error
	cursor   int
	width    int
	height   int
	focused  bool
	selected string
}

func newHomeModel(kc *kube.KubeconfigInfo, loadErr error) homeModel {
	m := homeModel{kc: kc, loadErr: loadErr}
	if kc != nil {
		for i, name := range kc.Contexts {
			if name == kc.CurrentContext {
				m.cursor = i
				break
			}
		}
	}
	return m
}

func (m homeModel) Init() tea.Cmd { return nil }

func (m homeModel) setSize(w, h int) homeModel {
	m.width = w
	m.height = h
	return m
}

func (m homeModel) focus() homeModel {
	m.focused = true
	return m
}

func (m homeModel) blur() homeModel {
	m.focused = false
	return m
}

// selection returns the chosen context name, empty if none yet.
func (m homeModel) selection() string { return m.selected }

func (m homeModel) Update(msg tea.Msg) (homeModel, bool, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); !ok {
		return m, false, nil
	}
	if !m.focused {
		return m, false, nil
	}
	km := msg.(tea.KeyPressMsg)
	if m.loadErr != nil || m.kc == nil || len(m.kc.Contexts) == 0 {
		return m, false, nil
	}
	switch {
	case key.Matches(km, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(km, keys.Down):
		if m.cursor < len(m.kc.Contexts)-1 {
			m.cursor++
		}
	case key.Matches(km, keys.Enter):
		m.selected = m.kc.Contexts[m.cursor]
		return m, true, nil
	}
	return m, false, nil
}

// centerBox draws content inside the standard bordered box, centered in the
// available terminal space (falling back to a sane default size).
func (m homeModel) centerBox(content string) string {
	boxed := styleBorder.Padding(1, 2).Render(content)
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, boxed)
}

func (m homeModel) viewLoading() string {
	title := styleTitle.Render("Clyde — Checking cluster")
	body := styleStatusVal.Render("Context: " + m.selected)
	spinner := styleHelp.Render("Verifying Goldmane availability…")
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", spinner)
	return m.centerBox(content)
}

// viewInstallPrompt asks the user whether to enable Goldmane/Whisker on a
// cluster that doesn't currently have it.
func (m homeModel) viewInstallPrompt() string {
	title := styleTitle.Render("Clyde — Goldmane / Whisker not enabled")
	body := lipgloss.JoinVertical(lipgloss.Left,
		styleStatusKey.Render("Context: ")+styleStatusVal.Render(m.selected),
		"",
		styleStatusVal.Render("Goldmane / Whisker is not enabled on this cluster."),
		styleStatusVal.Render("Clyde can enable it by creating the Goldmane and Whisker"),
		styleStatusVal.Render("resources via the installed Calico operator (requires Calico v3.30+)."),
	)
	footer := styleHelp.Render("y to install  ·  n/esc to cancel  ·  q to quit")
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", footer)
	return m.centerBox(content)
}

// viewInstalling shows progress while Goldmane/Whisker is being enabled.
func (m homeModel) viewInstalling() string {
	title := styleTitle.Render("Clyde — Enabling Goldmane / Whisker")
	body := styleStatusVal.Render("Context: " + m.selected)
	spinner := styleHelp.Render("Creating resources and waiting for Whisker to become available…")
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", spinner)
	return m.centerBox(content)
}

func (m homeModel) View() string {
	title := styleTitle.Render("Clyde — Select Kubernetes Context")

	var headerLines []string
	if m.kc != nil {
		headerLines = append(headerLines,
			styleStatusKey.Render("Kubeconfig: ")+styleStatusVal.Render(m.kc.Path),
			styleStatusKey.Render("Source:     ")+styleStatusVal.Render(m.kc.Source),
		)
		if m.kc.CurrentContext != "" {
			headerLines = append(headerLines,
				styleStatusKey.Render("Current:    ")+styleStatusVal.Render(m.kc.CurrentContext),
			)
		}
	}

	var body string
	switch {
	case m.loadErr != nil:
		body = styleError.Render(fmt.Sprintf("Failed to load kubeconfig: %v", m.loadErr))
	case m.kc == nil || len(m.kc.Contexts) == 0:
		body = styleError.Render("No contexts found in kubeconfig.")
	default:
		var lines []string
		for i, name := range m.kc.Contexts {
			marker := "  "
			if name == m.kc.CurrentContext {
				marker = "* "
			}
			line := marker + name
			if i == m.cursor {
				line = styleMenuItemSelected.Render(line)
			} else {
				line = styleMenuItem.Render(line)
			}
			lines = append(lines, line)
		}
		body = lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	footer := styleHelp.Render("↑/↓ to move  ·  enter to select  ·  q to quit")

	parts := []string{title, ""}
	parts = append(parts, headerLines...)
	parts = append(parts, "", body, "", footer)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return m.centerBox(content)
}
