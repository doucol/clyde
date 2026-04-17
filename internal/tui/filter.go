package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
)

const (
	fieldAction = iota
	fieldPort
	fieldNamespace
	fieldName
	fieldLabel
	fieldDateFrom
	fieldDateTo
	buttonSave
	buttonCancel
	buttonClear
	filterFieldCount = buttonClear + 1
)

var actionOptions = []string{"All", "Deny", "Allow"}

type filterModel struct {
	width     int
	height    int
	focusIdx  int
	inputs    []textinput.Model
	actionIdx int
	err       string
}

func newFilterModel() filterModel {
	current := global.GetFilter()
	inputs := make([]textinput.Model, fieldDateTo+1)

	for i := range inputs {
		t := textinput.New()
		t.Prompt = ""
		t.CharLimit = 60
		t.SetWidth(50)
		inputs[i] = t
	}

	inputs[fieldPort].CharLimit = 6
	inputs[fieldPort].SetWidth(10)
	inputs[fieldPort].Validate = func(s string) error {
		if s == "" {
			return nil
		}
		if _, err := strconv.Atoi(s); err != nil {
			return fmt.Errorf("must be numeric")
		}
		return nil
	}
	if current.Port > 0 {
		inputs[fieldPort].SetValue(strconv.Itoa(current.Port))
	}

	inputs[fieldNamespace].SetValue(current.Namespace)
	inputs[fieldName].SetValue(current.Name)
	inputs[fieldLabel].SetValue(current.Label)
	inputs[fieldDateFrom].SetValue(tf(current.DateFrom))
	inputs[fieldDateFrom].SetWidth(24)
	inputs[fieldDateFrom].Placeholder = time.RFC3339
	inputs[fieldDateTo].SetValue(tf(current.DateTo))
	inputs[fieldDateTo].SetWidth(24)
	inputs[fieldDateTo].Placeholder = time.RFC3339

	actionIdx := 0
	switch current.Action {
	case "Deny":
		actionIdx = 1
	case "Allow":
		actionIdx = 2
	}

	m := filterModel{
		inputs:    inputs,
		actionIdx: actionIdx,
		focusIdx:  fieldAction,
	}
	return m
}

func (m filterModel) Init() tea.Cmd { return textinput.Blink }

func (m filterModel) setSize(w, h int) filterModel {
	m.width = w
	m.height = h
	return m
}

type filterResult int

const (
	filterResultNone filterResult = iota
	filterResultSave
	filterResultCancel
	filterResultClear
)

func (m filterModel) Update(msg tea.Msg) (filterModel, filterResult, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Back):
			return m, filterResultCancel, nil
		case msg.String() == "tab" || msg.String() == "down":
			m.focusIdx = (m.focusIdx + 1) % filterFieldCount
			return m, filterResultNone, m.syncFocus()
		case msg.String() == "shift+tab" || msg.String() == "up":
			m.focusIdx = (m.focusIdx - 1 + filterFieldCount) % filterFieldCount
			return m, filterResultNone, m.syncFocus()
		case msg.String() == "enter":
			switch m.focusIdx {
			case buttonSave:
				return m, filterResultSave, nil
			case buttonCancel:
				return m, filterResultCancel, nil
			case buttonClear:
				return m, filterResultClear, nil
			default:
				return m, filterResultSave, nil
			}
		case msg.String() == "left", msg.String() == "right":
			if m.focusIdx == fieldAction {
				if msg.String() == "left" {
					m.actionIdx = (m.actionIdx - 1 + len(actionOptions)) % len(actionOptions)
				} else {
					m.actionIdx = (m.actionIdx + 1) % len(actionOptions)
				}
				return m, filterResultNone, nil
			}
		}
		if m.focusIdx >= fieldPort && m.focusIdx <= fieldDateTo {
			var cmd tea.Cmd
			m.inputs[m.focusIdx], cmd = m.inputs[m.focusIdx].Update(msg)
			return m, filterResultNone, cmd
		}
	}
	return m, filterResultNone, nil
}

func (m *filterModel) syncFocus() tea.Cmd {
	var cmds []tea.Cmd
	for i := fieldPort; i <= fieldDateTo; i++ {
		if i == m.focusIdx {
			cmds = append(cmds, m.inputs[i].Focus())
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m filterModel) toAttributes() (flowdata.FilterAttributes, error) {
	fa := flowdata.FilterAttributes{}
	if m.actionIdx > 0 {
		fa.Action = actionOptions[m.actionIdx]
	}
	if p := strings.TrimSpace(m.inputs[fieldPort].Value()); p != "" {
		port, err := strconv.Atoi(p)
		if err != nil {
			return fa, fmt.Errorf("port: %w", err)
		}
		fa.Port = port
	}
	fa.Namespace = strings.TrimSpace(m.inputs[fieldNamespace].Value())
	fa.Name = strings.TrimSpace(m.inputs[fieldName].Value())
	fa.Label = strings.TrimSpace(m.inputs[fieldLabel].Value())
	if s := strings.TrimSpace(m.inputs[fieldDateFrom].Value()); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return fa, fmt.Errorf("date from: %w", err)
		}
		fa.DateFrom = t
	}
	if s := strings.TrimSpace(m.inputs[fieldDateTo].Value()); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return fa, fmt.Errorf("date to: %w", err)
		}
		fa.DateTo = t
	}
	return fa, nil
}

func (m filterModel) View() string {
	title := styleOverlayTitle.Render("Filter Attributes")
	rows := []string{title, ""}
	rows = append(rows, m.renderAction())
	rows = append(rows, m.renderField("Port:", m.inputs[fieldPort].View(), m.focusIdx == fieldPort))
	rows = append(rows, m.renderField("Namespace:", m.inputs[fieldNamespace].View(), m.focusIdx == fieldNamespace))
	rows = append(rows, m.renderField("Name:", m.inputs[fieldName].View(), m.focusIdx == fieldName))
	rows = append(rows, m.renderField("Label:", m.inputs[fieldLabel].View(), m.focusIdx == fieldLabel))
	rows = append(rows, m.renderField("Date From:", m.inputs[fieldDateFrom].View(), m.focusIdx == fieldDateFrom))
	rows = append(rows, m.renderField("Date To:", m.inputs[fieldDateTo].View(), m.focusIdx == fieldDateTo))
	rows = append(rows, "")
	rows = append(rows, m.renderButtons())
	if m.err != "" {
		rows = append(rows, styleError.Render(m.err))
	}
	rows = append(rows, styleHelp.Render("tab/shift+tab: move  |  enter: save  |  esc: cancel  |  ←/→: action"))
	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	boxed := styleBorder.Padding(1, 2).Width(78).Render(content)
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, boxed)
}

func (m filterModel) renderAction() string {
	label := styleFormLabel.Render(padRight("Action:", 12))
	opts := make([]string, len(actionOptions))
	for i, opt := range actionOptions {
		style := styleFormField
		if i == m.actionIdx {
			style = styleSelected
		}
		opts[i] = style.Render(opt)
	}
	marker := "  "
	if m.focusIdx == fieldAction {
		marker = styleMenuKey.Render("▶ ")
	}
	return marker + label + strings.Join(opts, " ")
}

func (m filterModel) renderField(label, view string, focused bool) string {
	marker := "  "
	if focused {
		marker = styleMenuKey.Render("▶ ")
	}
	l := styleFormLabel.Render(padRight(label, 12))
	field := styleFormField.Render(view)
	if focused {
		field = styleFormFieldFocused.Render(view)
	}
	return marker + l + field
}

func (m filterModel) renderButtons() string {
	labels := []string{"Save", "Cancel", "Clear"}
	focusTargets := []int{buttonSave, buttonCancel, buttonClear}
	rendered := make([]string, len(labels))
	for i, label := range labels {
		style := styleButton
		if m.focusIdx == focusTargets[i] {
			style = styleButtonFocused
		}
		rendered[i] = style.Render(label)
	}
	return strings.Join(rendered, " ")
}
