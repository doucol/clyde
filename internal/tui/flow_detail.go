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
	vh := h - 12
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
	title := styleTitle.Render("Calico Flow Detail")
	header := m.renderHeader()
	body := m.viewport.View()
	status := styleHelp.Render("esc: back  |  ↑/↓: scroll")
	inner := lipgloss.JoinVertical(lipgloss.Left, title, header, body, status)
	w := m.width - 2
	if w < 10 {
		w = 10
	}
	return styleBorder.Width(w).Render(inner)
}

func (m flowDetailModel) renderHeader() string {
	fd := m.flow
	if fd == nil {
		return ""
	}
	kv := func(label, value string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			styleStatusKey.Render(label+": "),
			styleStatusVal.Render(value),
		)
	}
	line1 := lipgloss.JoinHorizontal(lipgloss.Top,
		kv("SRC", padRight(fmt.Sprintf("%s / %s", fd.SourceNamespace, fd.SourceName), 40)),
		"  ",
		kv("DST", padRight(fmt.Sprintf("%s / %s", fd.DestNamespace, fd.DestName), 40)),
	)
	line2 := lipgloss.JoinHorizontal(lipgloss.Top,
		kv("RPT/PROTO:PORT", padRight(fmt.Sprintf("%s / %s:%d", fd.Reporter, fd.Protocol, fd.DestPort), 30)),
		"  ",
		kv("START", padRight(tf(fd.StartTime), 22)),
		"  ",
		kv("END", padRight(tf(fd.EndTime), 22)),
	)
	line3 := lipgloss.JoinHorizontal(lipgloss.Top,
		kv("P I/O - B I/O", padRight(fmt.Sprintf("%d / %d - %d / %d", fd.PacketsIn, fd.PacketsOut, fd.BytesIn, fd.BytesOut), 30)),
		"  ",
		kv("ACTION", actionStyled(fd.Action)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, line1, line2, line3)
}
