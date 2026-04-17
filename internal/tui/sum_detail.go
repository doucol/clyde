package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/doucol/clyde/internal/flowdata"
)

type sumDetailModel struct {
	fds     *flowdata.FlowDataStore
	fc      dataProvider
	fas     *flowAppState
	table   table.Model
	flows   []*flowdata.FlowData
	sumID   int
	header  *flowdata.FlowSum
	width   int
	height  int
	focused bool
}

func sumDetailColumns() []table.Column {
	return []table.Column{
		{Title: "START TIME", Width: 22},
		{Title: "END TIME", Width: 22},
		{Title: "SRC LABELS", Width: 24},
		{Title: "DST LABELS", Width: 24},
		{Title: "REPORTER", Width: 10},
		{Title: "PACK IN", Width: 8},
		{Title: "PACK OUT", Width: 9},
		{Title: "BYTE IN", Width: 10},
		{Title: "BYTE OUT", Width: 10},
		{Title: "ACTION", Width: 8},
	}
}

func newSumDetailModel(fds *flowdata.FlowDataStore, fc dataProvider, fas *flowAppState) sumDetailModel {
	t := table.New(
		table.WithColumns(sumDetailColumns()),
		table.WithFocused(false),
	)
	t.SetStyles(tableStyles())
	return sumDetailModel{
		fds:   fds,
		fc:    fc,
		fas:   fas,
		table: t,
	}
}

func (m sumDetailModel) Init() tea.Cmd {
	return nil
}

func (m sumDetailModel) setSize(w, h int) sumDetailModel {
	m.width = w
	m.height = h
	tableWidth := w - 2
	m.table.SetWidth(tableWidth)
	m.table.SetColumns(scaleColumns(sumDetailColumns(), tableWidth))
	th := h - 12
	if th < 3 {
		th = 3
	}
	m.table.SetHeight(th)
	return m
}

func (m sumDetailModel) currentSumID() int {
	switch m.fas.lastHomePage {
	case pageSummaryTotalsName:
		return m.fas.sumID
	case pageSummaryRatesName:
		return m.fas.rateID
	}
	return 0
}

func (m sumDetailModel) focus() (sumDetailModel, tea.Cmd) {
	m.focused = true
	m.table.Focus()
	id := m.currentSumID()
	if id <= 0 {
		return m, nil
	}
	m.sumID = id
	m.header = m.fds.GetFlowSum(id)
	return m, fetchFlowsBySum(m.fc, id)
}

func (m sumDetailModel) blur() sumDetailModel {
	m.focused = false
	m.table.Blur()
	return m
}

func (m sumDetailModel) setFlows(flows []*flowdata.FlowData) sumDetailModel {
	m.flows = flows
	tableRows := make([]table.Row, len(flows))
	for i, fd := range flows {
		tableRows[i] = table.Row{
			tf(fd.StartTime),
			tf(fd.EndTime),
			fd.SourceLabels,
			fd.DestLabels,
			fd.Reporter,
			intos(fd.PacketsIn),
			intos(fd.PacketsOut),
			intos(fd.BytesIn),
			intos(fd.BytesOut),
			actionStyled(fd.Action),
		}
	}
	m.table.SetRows(tableRows)
	m.syncCursor()
	return m
}

func (m *sumDetailModel) syncCursor() {
	if len(m.flows) == 0 {
		m.fas.setFlow(0, 0)
		return
	}
	row := m.fas.flowRow
	if row == 0 {
		row = 1
	}
	cursor := row - 1
	if cursor >= len(m.flows) {
		cursor = len(m.flows) - 1
	}
	if cursor < 0 {
		cursor = 0
	}
	m.table.SetCursor(cursor)
	m.fas.setFlow(m.flows[cursor].ID, cursor+1)
}

func (m sumDetailModel) trackCursor() sumDetailModel {
	if len(m.flows) == 0 {
		m.fas.setFlow(0, 0)
		return m
	}
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.flows) {
		return m
	}
	m.fas.setFlow(m.flows[cursor].ID, cursor+1)
	return m
}

func (m sumDetailModel) Update(msg tea.Msg) (sumDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case flowsBySumMsg:
		if msg.sumID == m.sumID {
			m = m.setFlows(msg.flows)
		}
		return m, nil
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		if key.Matches(msg, m.table.KeyMap.LineUp, m.table.KeyMap.LineDown,
			m.table.KeyMap.PageUp, m.table.KeyMap.PageDown,
			m.table.KeyMap.HalfPageUp, m.table.KeyMap.HalfPageDown,
			m.table.KeyMap.GotoTop, m.table.KeyMap.GotoBottom) {
			m.table, _ = m.table.Update(msg)
			m = m.trackCursor()
			return m, nil
		}
	}
	return m, nil
}

func (m sumDetailModel) View() string {
	title := styleTitle.Render("Calico Flow Summary Detail")
	header := m.renderHeader()
	body := m.table.View()
	status := m.statusLine()
	inner := lipgloss.JoinVertical(lipgloss.Left, title, header, body, status)
	w := m.width - 2
	if w < 10 {
		w = 10
	}
	return styleBorder.Width(w).Render(inner)
}

func (m sumDetailModel) renderHeader() string {
	fs := m.header
	if fs == nil {
		return styleHelp.Render("(no summary selected)")
	}
	kv := func(label, value string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			styleStatusKey.Render(label+": "),
			styleStatusVal.Render(value),
		)
	}
	line1 := lipgloss.JoinHorizontal(lipgloss.Top,
		kv("SRC NS", padRight(fs.SourceNamespace, 20)),
		"  ",
		kv("SRC NAME", padRight(fs.SourceName, 30)),
	)
	line2 := lipgloss.JoinHorizontal(lipgloss.Top,
		kv("DST NS", padRight(fs.DestNamespace, 20)),
		"  ",
		kv("DST NAME", padRight(fs.DestName, 30)),
	)
	line3 := lipgloss.JoinHorizontal(lipgloss.Top,
		kv("PROTO", padRight(fs.Protocol, 8)),
		"  ",
		kv("PORT", fmt.Sprintf("%d", fs.DestPort)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, line1, line2, line3)
}

func (m sumDetailModel) statusLine() string {
	return styleHelp.Render(fmt.Sprintf("rows: %d  |  esc: back  |  enter: flow detail", len(m.flows)))
}

func padRight(s string, n int) string {
	runes := utf8.RuneCountInString(s)
	if runes >= n {
		return s
	}
	return s + strings.Repeat(" ", n-runes)
}
