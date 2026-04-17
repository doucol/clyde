package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
)

type summaryVariant int

const (
	variantTotals summaryVariant = iota
	variantRates
)

type summaryModel struct {
	variant variantProvider
	fc      dataProvider
	fas     *flowAppState
	table   table.Model
	rows    []*flowdata.FlowSum
	width   int
	height  int
	focused bool
}

type variantProvider interface {
	kind() summaryVariant
	title() string
	columns() []table.Column
	toRow(fs *flowdata.FlowSum) table.Row
	pageName() string
	onFocus(fas *flowAppState)
	setSelection(fas *flowAppState, id, row int)
	selectedRow(fas *flowAppState) int
	fetch(fc dataProvider) tea.Cmd
	msgType() string
}

type totalsVariant struct{}

func (totalsVariant) kind() summaryVariant { return variantTotals }
func (totalsVariant) title() string        { return "Calico Flow Summary Totals" }
func (totalsVariant) columns() []table.Column {
	return []table.Column{
		{Title: "SRC NAMESPACE / NAME", Width: 28},
		{Title: "DST NAMESPACE / NAME", Width: 28},
		{Title: "PROTO:PORT", Width: 12},
		{Title: "SRC / DST", Width: 10},
		{Title: "SRC PACK I/O", Width: 16},
		{Title: "SRC BYTE I/O", Width: 18},
		{Title: "DST PACK I/O", Width: 16},
		{Title: "DST BYTE I/O", Width: 18},
		{Title: "ACTION", Width: 8},
	}
}

func (totalsVariant) toRow(fs *flowdata.FlowSum) table.Row {
	return table.Row{
		fmt.Sprintf("%s / %s", fs.SourceNamespace, fs.SourceName),
		fmt.Sprintf("%s / %s", fs.DestNamespace, fs.DestName),
		fmt.Sprintf("%s:%d", fs.Protocol, fs.DestPort),
		fmt.Sprintf("%d / %d", fs.SourceReports, fs.DestReports),
		fmt.Sprintf("%d / %d", fs.SourcePacketsIn, fs.SourcePacketsOut),
		fmt.Sprintf("%d / %d", fs.SourceBytesIn, fs.SourceBytesOut),
		fmt.Sprintf("%d / %d", fs.DestPacketsIn, fs.DestPacketsOut),
		fmt.Sprintf("%d / %d", fs.DestBytesIn, fs.DestBytesOut),
		actionStyled(fs.Action),
	}
}

func (totalsVariant) pageName() string { return pageSummaryTotalsName }

func (totalsVariant) onFocus(fas *flowAppState) {
	fas.lastHomePage = pageSummaryTotalsName
}

func (totalsVariant) setSelection(fas *flowAppState, id, row int) {
	fas.setSum(id, row)
}

func (totalsVariant) selectedRow(fas *flowAppState) int { return fas.sumRow }

func (totalsVariant) fetch(fc dataProvider) tea.Cmd { return fetchSumTotals(fc) }

func (totalsVariant) msgType() string { return "totals" }

type ratesVariant struct{}

func (ratesVariant) kind() summaryVariant { return variantRates }
func (ratesVariant) title() string        { return "Calico Flow Summary Rates" }
func (ratesVariant) columns() []table.Column {
	return []table.Column{
		{Title: "SRC NAMESPACE / NAME", Width: 28},
		{Title: "DST NAMESPACE / NAME", Width: 28},
		{Title: "PROTO:PORT", Width: 12},
		{Title: "SRC PACK/SEC", Width: 14},
		{Title: "SRC BYTE/SEC", Width: 14},
		{Title: "DST PACK/SEC", Width: 14},
		{Title: "DST BYTE/SEC", Width: 14},
		{Title: "ACTION", Width: 8},
	}
}

func (ratesVariant) toRow(fs *flowdata.FlowSum) table.Row {
	return table.Row{
		fmt.Sprintf("%s / %s", fs.SourceNamespace, fs.SourceName),
		fmt.Sprintf("%s / %s", fs.DestNamespace, fs.DestName),
		fmt.Sprintf("%s:%d", fs.Protocol, fs.DestPort),
		fmt.Sprintf("%.2f", fs.SourceTotalPacketRate),
		fmt.Sprintf("%.2f", fs.SourceTotalByteRate),
		fmt.Sprintf("%.2f", fs.DestTotalPacketRate),
		fmt.Sprintf("%.2f", fs.DestTotalByteRate),
		actionStyled(fs.Action),
	}
}

func (ratesVariant) pageName() string { return pageSummaryRatesName }

func (ratesVariant) onFocus(fas *flowAppState) {
	fas.lastHomePage = pageSummaryRatesName
}

func (ratesVariant) setSelection(fas *flowAppState, id, row int) {
	fas.setRate(id, row)
}

func (ratesVariant) selectedRow(fas *flowAppState) int { return fas.rateRow }

func (ratesVariant) fetch(fc dataProvider) tea.Cmd { return fetchSumRates(fc) }

func (ratesVariant) msgType() string { return "rates" }

func newSummaryModel(v variantProvider, fc dataProvider, fas *flowAppState) summaryModel {
	t := table.New(
		table.WithColumns(v.columns()),
		table.WithFocused(false),
	)
	t.SetStyles(tableStyles())
	return summaryModel{
		variant: v,
		fc:      fc,
		fas:     fas,
		table:   t,
	}
}

func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = styleHeaderCell
	s.Cell = styleCell
	s.Selected = styleSelected
	return s
}

func (m summaryModel) Init() tea.Cmd {
	return m.variant.fetch(m.fc)
}

func (m summaryModel) setSize(w, h int) summaryModel {
	m.width = w
	m.height = h
	m.table.SetWidth(w - 2)
	th := h - 4
	if th < 3 {
		th = 3
	}
	m.table.SetHeight(th)
	return m
}

func (m summaryModel) focus() summaryModel {
	m.focused = true
	m.table.Focus()
	m.variant.onFocus(m.fas)
	return m
}

func (m summaryModel) blur() summaryModel {
	m.focused = false
	m.table.Blur()
	return m
}

func (m summaryModel) setRows(rows []*flowdata.FlowSum) summaryModel {
	m.rows = rows
	tableRows := make([]table.Row, len(rows))
	for i, fs := range rows {
		tableRows[i] = m.variant.toRow(fs)
	}
	m.table.SetRows(tableRows)
	m.syncCursor()
	return m
}

func (m *summaryModel) syncCursor() {
	if len(m.rows) == 0 {
		m.variant.setSelection(m.fas, 0, 0)
		return
	}
	row := m.variant.selectedRow(m.fas)
	if row == 0 {
		row = 1
	}
	cursor := row - 1
	if cursor >= len(m.rows) {
		cursor = len(m.rows) - 1
	}
	if cursor < 0 {
		cursor = 0
	}
	m.table.SetCursor(cursor)
	m.variant.setSelection(m.fas, m.rows[cursor].ID, cursor+1)
}

func (m summaryModel) trackCursor() summaryModel {
	if len(m.rows) == 0 {
		m.variant.setSelection(m.fas, 0, 0)
		return m
	}
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.rows) {
		return m
	}
	m.variant.setSelection(m.fas, m.rows[cursor].ID, cursor+1)
	return m
}

func (m summaryModel) Update(msg tea.Msg) (summaryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case flowSumTotalsMsg:
		if m.variant.kind() == variantTotals {
			m = m.setRows([]*flowdata.FlowSum(msg))
		}
		return m, nil
	case flowSumRatesMsg:
		if m.variant.kind() == variantRates {
			m = m.setRows([]*flowdata.FlowSum(msg))
		}
		return m, nil
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		switch {
		case key.Matches(msg, keys.SortKey):
			return m, m.toggleSort("Key", true)
		case key.Matches(msg, keys.SortSrcPkt):
			if m.variant.kind() == variantRates {
				return m, m.toggleSort("SourceTotalPacketRate", false)
			}
		case key.Matches(msg, keys.SortDstPkt):
			if m.variant.kind() == variantRates {
				return m, m.toggleSort("DestTotalPacketRate", false)
			}
		case key.Matches(msg, keys.SortSrcByte):
			if m.variant.kind() == variantRates {
				return m, m.toggleSort("SourceTotalByteRate", false)
			}
		case key.Matches(msg, keys.SortDstByte):
			if m.variant.kind() == variantRates {
				return m, m.toggleSort("DestTotalByteRate", false)
			}
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

func (m summaryModel) toggleSort(fieldName string, defaultAsc bool) tea.Cmd {
	sa := global.GetSort()
	asc := defaultAsc
	switch m.variant.kind() {
	case variantTotals:
		if sa.SumTotalsFieldName == fieldName {
			asc = !sa.SumTotalsAscending
		}
		global.SetSort(flowdata.SortAttributes{
			SumTotalsFieldName: fieldName,
			SumTotalsAscending: asc,
			SumRatesFieldName:  sa.SumRatesFieldName,
			SumRatesAscending:  sa.SumRatesAscending,
		})
		m.fas.setSum(0, 0)
	case variantRates:
		if sa.SumRatesFieldName == fieldName {
			asc = !sa.SumRatesAscending
		}
		global.SetSort(flowdata.SortAttributes{
			SumTotalsFieldName: sa.SumTotalsFieldName,
			SumTotalsAscending: sa.SumTotalsAscending,
			SumRatesFieldName:  fieldName,
			SumRatesAscending:  asc,
		})
		m.fas.setRate(0, 0)
	}
	return m.variant.fetch(m.fc)
}

func (m summaryModel) View() string {
	title := styleTitle.Render(m.variant.title())
	body := m.table.View()
	status := m.statusLine()
	inner := lipgloss.JoinVertical(lipgloss.Left, title, body, status)
	w := m.width - 2
	if w < 10 {
		w = 10
	}
	return styleBorder.Width(w).Render(inner)
}

func (m summaryModel) statusLine() string {
	sa := global.GetSort()
	var sortText string
	switch m.variant.kind() {
	case variantTotals:
		if sa.SumTotalsFieldName != "" {
			sortText = fmt.Sprintf("sort: %s %s", sa.SumTotalsFieldName, ascDesc(sa.SumTotalsAscending))
		}
	case variantRates:
		if sa.SumRatesFieldName != "" {
			sortText = fmt.Sprintf("sort: %s %s", sa.SumRatesFieldName, ascDesc(sa.SumRatesAscending))
		}
	}
	filterText := ""
	f := global.GetFilter()
	if f != (flowdata.FilterAttributes{}) {
		filterText = "filter: on"
	}
	count := fmt.Sprintf("rows: %d", len(m.rows))
	return styleHelp.Render(joinStatus(count, sortText, filterText))
}

func ascDesc(asc bool) string {
	if asc {
		return "asc"
	}
	return "desc"
}

func joinStatus(parts ...string) string {
	out := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		if out != "" {
			out += "  |  "
		}
		out += p
	}
	return out
}
