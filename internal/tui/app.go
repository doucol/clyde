// Package tui provides the terminal user interface for Clyde.
package tui

import (
	"context"
	"errors"
	"sync"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
	"github.com/doucol/clyde/internal/util"
)

// ErrGoldmaneNotAvailable is returned from FlowApp.Run when the selected
// context does not have the Calico Goldmane log aggregator available.
var ErrGoldmaneNotAvailable = errors.New(
	"Goldmane is not available on the selected context. " +
		"Clyde requires Calico v3.30+ with the Goldmane log aggregator enabled. " +
		"See https://docs.tigera.io/calico/latest/observability/goldmane for upgrade instructions.",
)

const (
	pageHomeName          = "home"
	pageSummaryTotalsName = "summaryTotals"
	pageSummaryRatesName  = "summaryRates"
	pageSumDetailName     = "sumDetail"
	pageFlowDetailName    = "flowDetail"
)

type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayHelp
	overlayFilter
)

type FlowApp struct {
	mu      *sync.Mutex
	fds     *flowdata.FlowDataStore
	fc      *flowcache.FlowCache
	fas     *flowAppState
	pages   pageRegistry
	prog    *tea.Program
	exitErr error
}

type pageRegistry struct {
	active string
}

func NewFlowApp(fds *flowdata.FlowDataStore, fc *flowcache.FlowCache) *FlowApp {
	return &FlowApp{
		mu:    &sync.Mutex{},
		fds:   fds,
		fc:    fc,
		fas:   &flowAppState{},
		pages: pageRegistry{active: pageHomeName},
	}
}

type appModel struct {
	fa *FlowApp

	width, height int

	page    string
	overlay overlayKind

	home       homeModel
	totals     summaryModel
	rates      summaryModel
	sumDetail  sumDetailModel
	flowDetail flowDetailModel

	help    helpModel
	filter  filterModel
	loading bool // goldmane check in flight

	ctx context.Context
	cc  *cmdctx.CmdCtx
}

func (fa *FlowApp) newAppModel(ctx context.Context) appModel {
	cc := cmdctx.CmdCtxFromContext(ctx)
	kc, loadErr := util.LoadKubeconfigInfo(cc.KubeconfigPath(), cc.KubeconfigSource())

	m := appModel{
		fa:         fa,
		page:       pageHomeName,
		home:       newHomeModel(kc, loadErr).focus(),
		totals:     newSummaryModel(totalsVariant{}, fa.fc, fa.fas),
		rates:      newSummaryModel(ratesVariant{}, fa.fc, fa.fas),
		sumDetail:  newSumDetailModel(fa.fds, fa.fc, fa.fas),
		flowDetail: newFlowDetailModel(fa.fds, fa.fas),
		help:       newHelpModel(),
		filter:     newFilterModel(),
		ctx:        ctx,
		cc:         cc,
	}
	return m
}

// initialAutoSelect returns the context to auto-select (skipping the picker),
// or empty string if the picker should be shown.
func (m appModel) initialAutoSelect() string {
	if m.cc.KubeContext() != "" {
		return m.cc.KubeContext()
	}
	if m.home.kc != nil && len(m.home.kc.Contexts) == 1 {
		return m.home.kc.Contexts[0]
	}
	return ""
}

type autoSelectMsg struct{ name string }

func (m appModel) Init() tea.Cmd {
	cmds := []tea.Cmd{tickCmd(), m.home.Init()}
	if sel := m.initialAutoSelect(); sel != "" {
		name := sel
		cmds = append(cmds, func() tea.Msg { return autoSelectMsg{name: name} })
	}
	return tea.Batch(cmds...)
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.propagateSize()
		return m, nil

	case tickMsg:
		return m, tea.Batch(tickCmd(), m.refreshCmd())

	case flowSumTotalsMsg, flowSumRatesMsg:
		var cmd1, cmd2 tea.Cmd
		m.totals, cmd1 = m.totals.Update(msg)
		m.rates, cmd2 = m.rates.Update(msg)
		return m, tea.Batch(cmd1, cmd2)

	case flowsBySumMsg:
		var cmd tea.Cmd
		m.sumDetail, cmd = m.sumDetail.Update(msg)
		return m, cmd

	case autoSelectMsg:
		return m.onContextSelected(msg.name, nil)

	case clusterReadyMsg:
		m.loading = false
		if !msg.info.WhiskerAvailable {
			m.fa.setExitErr(ErrGoldmaneNotAvailable)
			return m, tea.Quit
		}
		return m.gotoPage(pageSummaryTotalsName)

	case tea.KeyPressMsg:
		if m.overlay != overlayNone {
			return m.updateOverlay(msg)
		}
		return m.updatePage(msg)
	}
	return m, nil
}

func (m appModel) updateOverlay(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.overlay {
	case overlayHelp:
		var close bool
		m.help, close = m.help.Update(msg)
		if close {
			m.overlay = overlayNone
		}
		return m, nil
	case overlayFilter:
		var result filterResult
		var cmd tea.Cmd
		m.filter, result, cmd = m.filter.Update(msg)
		switch result {
		case filterResultCancel:
			m.overlay = overlayNone
			return m, cmd
		case filterResultSave:
			fa, err := m.filter.toAttributes()
			if err != nil {
				m.filter.err = err.Error()
				return m, cmd
			}
			if fa != global.GetFilter() {
				global.SetFilter(fa)
				m.fa.fas.reset()
			}
			m.overlay = overlayNone
			target := m.fa.fas.lastHomePage
			if target == "" {
				target = pageSummaryTotalsName
			}
			return m.gotoPage(target)
		case filterResultClear:
			global.SetFilter(flowdata.FilterAttributes{})
			m.fa.fas.reset()
			m.overlay = overlayNone
			target := m.fa.fas.lastHomePage
			if target == "" {
				target = pageSummaryTotalsName
			}
			return m.gotoPage(target)
		}
		return m, cmd
	}
	return m, nil
}

func (m appModel) updatePage(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, m.quitCmd()
	case key.Matches(msg, keys.Help):
		m.help = m.help.setSize(m.width, m.height)
		m.overlay = overlayHelp
		return m, nil
	case key.Matches(msg, keys.Filter):
		if m.page == pageHomeName {
			return m, nil
		}
		m.filter = newFilterModel().setSize(m.width, m.height)
		m.overlay = overlayFilter
		return m, nil
	case key.Matches(msg, keys.Home):
		if m.page != pageHomeName {
			return m.gotoPage(pageHomeName)
		}
	case key.Matches(msg, keys.Rates):
		if m.page == pageSummaryTotalsName {
			return m.gotoPage(pageSummaryRatesName)
		}
	case key.Matches(msg, keys.Totals):
		if m.page == pageSummaryRatesName {
			return m.gotoPage(pageSummaryTotalsName)
		}
	case key.Matches(msg, keys.Back):
		return m.handleBack()
	}

	switch m.page {
	case pageHomeName:
		if m.loading {
			return m, nil
		}
		newHome, selected, cmd := m.home.Update(msg)
		m.home = newHome
		if selected {
			return m.onContextSelected(newHome.selection(), cmd)
		}
		return m, cmd
	case pageSummaryTotalsName:
		if key.Matches(msg, keys.Enter) {
			if m.fa.fas.sumRow > 0 {
				return m.gotoPage(pageSumDetailName)
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.totals, cmd = m.totals.Update(msg)
		return m, cmd
	case pageSummaryRatesName:
		if key.Matches(msg, keys.Enter) {
			if m.fa.fas.rateRow > 0 {
				return m.gotoPage(pageSumDetailName)
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.rates, cmd = m.rates.Update(msg)
		return m, cmd
	case pageSumDetailName:
		if key.Matches(msg, keys.Enter) {
			if m.fa.fas.flowRow > 0 {
				return m.gotoPage(pageFlowDetailName)
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.sumDetail, cmd = m.sumDetail.Update(msg)
		return m, cmd
	case pageFlowDetailName:
		var cmd tea.Cmd
		m.flowDetail, cmd = m.flowDetail.Update(msg)
		return m, cmd
	}

	return m, nil
}

// onContextSelected is called when the user picks a context in the home page.
func (m appModel) onContextSelected(name string, extra tea.Cmd) (tea.Model, tea.Cmd) {
	if name == "" {
		return m, extra
	}
	m.cc.SetContext(name)
	m.home.selected = name
	m.loading = true
	return m, tea.Batch(extra, checkClusterReadyCmd(m.ctx))
}

func (m appModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.page {
	case pageSummaryTotalsName, pageSummaryRatesName:
		return m.gotoPage(pageHomeName)
	case pageSumDetailName:
		target := m.fa.fas.lastHomePage
		if target == "" {
			target = pageSummaryTotalsName
		}
		return m.gotoPage(target)
	case pageFlowDetailName:
		return m.gotoPage(pageSumDetailName)
	}
	return m, nil
}

func (m appModel) quitCmd() tea.Cmd {
	return tea.Quit
}

func (m appModel) gotoPage(target string) (tea.Model, tea.Cmd) {
	prev := m.page
	m.page = target

	switch prev {
	case pageHomeName:
		m.home = m.home.blur()
	case pageSummaryTotalsName:
		m.totals = m.totals.blur()
	case pageSummaryRatesName:
		m.rates = m.rates.blur()
	case pageSumDetailName:
		m.sumDetail = m.sumDetail.blur()
	case pageFlowDetailName:
		m.flowDetail = m.flowDetail.blur()
	}

	var cmd tea.Cmd
	switch target {
	case pageHomeName:
		m.home = m.home.focus()
	case pageSummaryTotalsName:
		m.totals = m.totals.focus()
		cmd = fetchSumTotals(m.fa.fc)
	case pageSummaryRatesName:
		m.rates = m.rates.focus()
		cmd = fetchSumRates(m.fa.fc)
	case pageSumDetailName:
		var focusCmd tea.Cmd
		m.sumDetail, focusCmd = m.sumDetail.focus()
		cmd = focusCmd
	case pageFlowDetailName:
		m.flowDetail = m.flowDetail.focus()
	}
	return m, cmd
}

func (m appModel) propagateSize() appModel {
	m.home = m.home.setSize(m.width, m.height)
	m.totals = m.totals.setSize(m.width, m.height)
	m.rates = m.rates.setSize(m.width, m.height)
	m.sumDetail = m.sumDetail.setSize(m.width, m.height)
	m.flowDetail = m.flowDetail.setSize(m.width, m.height)
	m.help = m.help.setSize(m.width, m.height)
	m.filter = m.filter.setSize(m.width, m.height)
	return m
}

func (m appModel) refreshCmd() tea.Cmd {
	switch m.page {
	case pageSummaryTotalsName:
		return fetchSumTotals(m.fa.fc)
	case pageSummaryRatesName:
		return fetchSumRates(m.fa.fc)
	case pageSumDetailName:
		id := m.sumDetail.currentSumID()
		if id > 0 {
			return fetchFlowsBySum(m.fa.fc, id)
		}
	}
	return nil
}

func (m appModel) View() tea.View {
	var body string
	switch m.page {
	case pageHomeName:
		if m.loading {
			body = m.home.viewLoading()
		} else {
			body = m.home.View()
		}
	case pageSummaryTotalsName:
		body = m.totals.View()
	case pageSummaryRatesName:
		body = m.rates.View()
	case pageSumDetailName:
		body = m.sumDetail.View()
	case pageFlowDetailName:
		body = m.flowDetail.View()
	}

	var overlay string
	switch m.overlay {
	case overlayHelp:
		overlay = m.help.View()
	case overlayFilter:
		overlay = m.filter.View()
	}

	content := body
	if overlay != "" && m.width > 0 && m.height > 0 {
		ow := lipgloss.Width(overlay)
		oh := lipgloss.Height(overlay)
		ox := max((m.width-ow)/2, 0)
		oy := max((m.height-oh)/2, 0)
		canvas := lipgloss.NewCanvas(m.width, m.height)
		canvas.Compose(lipgloss.NewCompositor(
			lipgloss.NewLayer(body),
			lipgloss.NewLayer(overlay).X(ox).Y(oy).Z(1),
		))
		content = canvas.Render()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (fa *FlowApp) setExitErr(err error) {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	fa.exitErr = err
}

func (fa *FlowApp) ExitErr() error {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	return fa.exitErr
}

func (fa *FlowApp) Run(ctx context.Context) error {
	model := fa.newAppModel(ctx)
	prog := tea.NewProgram(model, tea.WithContext(ctx))
	fa.mu.Lock()
	fa.prog = prog
	fa.mu.Unlock()

	_, err := prog.Run()
	// Cancel the command context so sibling goroutines (flow catcher, etc.)
	// can shut down once the TUI exits. Runs regardless of how we got here.
	if cc := cmdctx.CmdCtxFromContext(ctx); cc != nil {
		cc.Cancel()
	}
	// A cancelled context — either via our own Cancel above or an external
	// signal — is a normal shutdown, not a program error.
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return fa.ExitErr()
}

func (fa *FlowApp) Stop() {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	if fa.prog != nil {
		fa.prog.Quit()
		fa.prog = nil
	}
}

// updateSort preserved for tests. New paths go through summaryModel.toggleSort.
func (fa *FlowApp) updateSort(_ any, fieldName string, defaultOrder bool, pageName string) any {
	sa := global.GetSort()
	asc := defaultOrder
	switch pageName {
	case pageSummaryTotalsName:
		if sa.SumTotalsFieldName == fieldName {
			asc = !sa.SumTotalsAscending
		}
		global.SetSort(flowdata.SortAttributes{
			SumTotalsFieldName: fieldName,
			SumTotalsAscending: asc,
			SumRatesFieldName:  sa.SumRatesFieldName,
			SumRatesAscending:  sa.SumRatesAscending,
		})
		fa.fas.setSum(0, 0)
		return nil
	case pageSummaryRatesName:
		if sa.SumRatesFieldName == fieldName {
			asc = !sa.SumRatesAscending
		}
		global.SetSort(flowdata.SortAttributes{
			SumTotalsFieldName: sa.SumTotalsFieldName,
			SumTotalsAscending: sa.SumTotalsAscending,
			SumRatesFieldName:  fieldName,
			SumRatesAscending:  asc,
		})
		fa.fas.setRate(0, 0)
		return nil
	}
	return "eventPassthrough"
}
