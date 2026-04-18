package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	colorBg         = lipgloss.Color("#000000")
	colorText       = lipgloss.Color("#D3D3D3")
	colorTitle      = lipgloss.Color("#FFFFFF")
	colorBorder     = lipgloss.Color("#2F4F4F")
	colorAllow      = lipgloss.Color("#87CEFA")
	colorDeny       = lipgloss.Color("#FF4500")
	colorSelBg      = lipgloss.Color("#00008B")
	colorSelFg      = lipgloss.Color("#FFFFFF")
	colorLabel      = lipgloss.Color("#E0E0E0")
	colorDim        = lipgloss.Color("#808080")
	colorAccent     = lipgloss.Color("#FFD700")
	colorError      = lipgloss.Color("#FF4500")
	colorSelectable = lipgloss.Color("#00CED1")
)

var (
	styleTitle = lipgloss.NewStyle().
			Foreground(colorTitle).
			Bold(true)

	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Background(colorBg)

	styleHeaderCell = lipgloss.NewStyle().
			Foreground(colorTitle).
			Background(colorBg).
			Bold(true).
			Padding(0, 1)

	styleCell = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorBg).
			Padding(0, 1)

	styleSelected = lipgloss.NewStyle().
			Foreground(colorSelFg).
			Background(colorSelBg).
			Bold(true).
			Padding(0, 1)

	styleAllow = lipgloss.NewStyle().Foreground(colorAllow).Bold(true)
	styleDeny  = lipgloss.NewStyle().Foreground(colorDeny).Bold(true)

	styleHelp = lipgloss.NewStyle().Foreground(colorDim)

	styleStatusKey = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleStatusVal = lipgloss.NewStyle().Foreground(colorText)

	styleFormLabel = lipgloss.NewStyle().
			Foreground(colorLabel).
			Bold(true)

	styleFormField = lipgloss.NewStyle().
			Foreground(colorSelFg).
			Background(colorSelBg).
			Padding(0, 1)

	styleFormFieldFocused = styleFormField.
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorAccent)

	styleButton = lipgloss.NewStyle().
			Foreground(colorTitle).
			Background(colorBorder).
			Padding(0, 2).
			MarginRight(2)

	styleButtonFocused = lipgloss.NewStyle().
				Foreground(colorSelFg).
				Background(colorSelBg).
				Bold(true).
				Padding(0, 2).
				MarginRight(2)

	styleMenuItem = lipgloss.NewStyle().
			Foreground(colorText).
			Padding(0, 2)

	styleMenuItemSelected = lipgloss.NewStyle().
				Foreground(colorAccent).
				Background(colorSelBg).
				Bold(true).
				Padding(0, 2)

	styleMenuKey = lipgloss.NewStyle().
			Foreground(colorSelectable).
			Bold(true)

	styleError = lipgloss.NewStyle().Foreground(colorError).Bold(true)
)

// renderTitledBorder wraps content in the rounded border and embeds the
// given title at the top-center, replacing the middle of the top border.
func renderTitledBorder(title, content string, width int) string {
	boxed := styleBorder.Width(width).Render(content)
	if title == "" {
		return boxed
	}
	lines := strings.Split(boxed, "\n")
	if len(lines) == 0 {
		return boxed
	}
	totalWidth := lipgloss.Width(lines[0])
	innerSpace := totalWidth - 2
	if innerSpace < 4 {
		return boxed
	}
	styledTitle := styleTitle.Render(" " + title + " ")
	titleW := lipgloss.Width(styledTitle)
	if titleW > innerSpace-2 {
		return boxed
	}
	leftDashes := (innerSpace - titleW) / 2
	rightDashes := innerSpace - titleW - leftDashes
	bs := lipgloss.NewStyle().Foreground(colorBorder)
	lines[0] = bs.Render("╭"+strings.Repeat("─", leftDashes)) +
		styledTitle +
		bs.Render(strings.Repeat("─", rightDashes)+"╮")
	return strings.Join(lines, "\n")
}

func actionStyled(action string) string {
	if action == "Deny" || action == "deny" {
		return styleDeny.Render(action)
	}
	return styleAllow.Render(action)
}
