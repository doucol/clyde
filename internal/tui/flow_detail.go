package tui

import (
	"fmt"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/doucol/clyde/internal/flowdata"
)

type flowDetailModel struct {
	fds      *flowdata.FlowDataStore
	fas      *flowAppState
	viewport viewport.Model
	flow     *flowdata.FlowData
	flowID   int
	width    int
	height   int
	focused  bool
}

func newFlowDetailModel(fds *flowdata.FlowDataStore, fas *flowAppState) flowDetailModel {
	vp := viewport.New()
	return flowDetailModel{
		fds:      fds,
		fas:      fas,
		viewport: vp,
	}
}

func (m flowDetailModel) Init() tea.Cmd {
	return nil
}

func (m flowDetailModel) setSize(w, h int) flowDetailModel {
	m.width = w
	m.height = h
	m.viewport.SetWidth(w - 4)
	// 2 border lines + 11 info-table lines (9 rows + 2 borders) + 1 status line
	vh := h - 14
	if vh < 3 {
		vh = 3
	}
	m.viewport.SetHeight(vh)
	m.refreshContent()
	return m
}

func (m flowDetailModel) focus() flowDetailModel {
	m.focused = true
	m.flowID = m.fas.flowID
	if m.flowID > 0 {
		m.flow = m.fds.GetFlowDetail(m.flowID)
	} else {
		m.flow = nil
	}
	m.refreshContent()
	return m
}

func (m flowDetailModel) blur() flowDetailModel {
	m.focused = false
	return m
}

func (m *flowDetailModel) refreshContent() {
	if m.flow == nil {
		m.viewport.SetContent(styleHelp.Render("(no flow selected)"))
		return
	}
	fd := m.flow
	body := fmt.Sprintf("SRC LABELS: %s\n\nDST LABELS: %s\n\nPolicy Hits Enforced:\n\n%sPolicy Hits Pending:\n\n%s",
		fd.SourceLabels, fd.DestLabels,
		policyHitsToString(fd.Policies.Enforced),
		policyHitsToString(fd.Policies.Pending))
	m.viewport.SetContent(body)
}

func (m flowDetailModel) Update(msg tea.Msg) (flowDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	case tea.MouseWheelMsg:
		if !m.focused {
			return m, nil
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m flowDetailModel) View() string {
	header := m.renderHeader()
	body := m.viewport.View()
	status := styleHelp.Render("esc: back  |  ↑/↓: scroll")
	inner := lipgloss.JoinVertical(lipgloss.Left, header, body, status)
	w := m.width - 2
	if w < 10 {
		w = 10
	}
	return renderTitledBorder("Calico Flow Detail", inner, w)
}

func (m flowDetailModel) renderHeader() string {
	fd := m.flow
	if fd == nil {
		return ""
	}
	return infoTable([][]string{
		{"Source", fmt.Sprintf("%s / %s", fd.SourceNamespace, fd.SourceName)},
		{"Destination", fmt.Sprintf("%s / %s", fd.DestNamespace, fd.DestName)},
		{"Reporter", fd.Reporter},
		{"Protocol:Port", fmt.Sprintf("%s:%d", fd.Protocol, fd.DestPort)},
		{"Started", tf(fd.StartTime)},
		{"Ended", tf(fd.EndTime)},
		{"Packets in/out", fmt.Sprintf("%d / %d", fd.PacketsIn, fd.PacketsOut)},
		{"Bytes in/out", fmt.Sprintf("%d / %d", fd.BytesIn, fd.BytesOut)},
		{"Action", actionStyled(fd.Action)},
	})
}
