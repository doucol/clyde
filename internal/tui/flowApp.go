package tui

import (
	"context"
	"sync"
	"time"

	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

const (
	pageSummaryTotalsName = "summaryTotals"
	pageSummaryRatesName  = "summaryRates"
	pageSumDetailName     = "sumDetail"
	pageFlowDetailName    = "flowDetail"
)

type FlowApp struct {
	mu    *sync.Mutex
	app   *tview.Application
	fds   *flowdata.FlowDataStore
	fc    *flowcache.FlowCache
	fas   *flowAppState
	pages *tview.Pages
}

func NewFlowApp(fds *flowdata.FlowDataStore, fc *flowcache.FlowCache) *FlowApp {
	setTheme()
	pages := tview.NewPages()
	pages.SetBackgroundColor(bgColor).SetBorderColor(borderColor).SetTitleColor(titleColor)
	pages.SetBorderStyle(tcell.StyleDefault.Foreground(borderColor).Background(bgColor))
	fas := &flowAppState{}
	app := tview.NewApplication()
	return &FlowApp{&sync.Mutex{}, app, fds, fc, fas, pages}
}

func (fa *FlowApp) updateSort(event *tcell.EventKey, fieldName string, defaultOrder bool, pageName string) *tcell.EventKey {
	if page, _ := fa.pages.GetFrontPage(); page == pageName {
		sa := global.GetSort()
		asc := defaultOrder
		if pageName == pageSummaryTotalsName {
			if sa.SumTotalsFieldName == fieldName {
				asc = !sa.SumTotalsAscending
			}
			global.SetSort(flowdata.SortAttributes{SumTotalsFieldName: fieldName, SumTotalsAscending: asc})
			fa.fas.setSum(0, 0)
			return nil
		} else if pageName == pageSummaryRatesName {
			if sa.SumRatesFieldName == fieldName {
				asc = !sa.SumRatesAscending
			}
			global.SetSort(flowdata.SortAttributes{SumRatesFieldName: fieldName, SumRatesAscending: asc})
			fa.fas.setRate(0, 0)
			return nil
		}
	}
	return event
}

func (fa *FlowApp) Run(ctx context.Context) error {
	defer fa.Stop()
	cc := cmdctx.CmdCtxFromContext(ctx)

	// Set up an input capture to shutdown the app when the user presses Ctrl-C
	fa.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		stop := func() *tcell.EventKey {
			fa.Stop()
			cc.Cancel()
			return nil
		}

		switch event.Key() {
		case tcell.KeyCtrlC:
			return stop()
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				return stop()
			case 'r':
				if page, _ := fa.pages.GetFrontPage(); page == pageSummaryTotalsName {
					fa.pages.SwitchToPage(pageSummaryRatesName)
					return nil
				}
			case 't':
				if page, _ := fa.pages.GetFrontPage(); page == pageSummaryRatesName {
					fa.pages.SwitchToPage(pageSummaryTotalsName)
					return nil
				}
			case 'p':
				return fa.updateSort(event, "SourceTotalPacketRate", false, pageSummaryRatesName)
			case 'P':
				return fa.updateSort(event, "DestTotalPacketRate", false, pageSummaryRatesName)
			case 'b':
				return fa.updateSort(event, "SourceTotalByteRate", false, pageSummaryRatesName)
			case 'B':
				return fa.updateSort(event, "DestTotalByteRate", false, pageSummaryRatesName)
			case 'n':
				page, _ := fa.pages.GetFrontPage()
				return fa.updateSort(event, "Key", true, page)
			case '/':
				if !fa.pages.HasPage(modalName) {
					fa.filterModal()
					return nil
				}
			}
		}
		return event
	})

	// Go update screen periodically until we're shutdown
	go func() {
		ticker := time.Tick(2 * time.Second)
		for {
			select {
			case <-ctx.Done():
				log.Debugf("flowApp shutting down: done signal received")
				fa.Stop()
				return
			case <-ticker:
				fa.app.Draw()
			}
		}
	}()

	fa.pages.AddPage(pageSummaryTotalsName, fa.viewSummary(), true, true)
	fa.pages.AddPage(pageSummaryRatesName, fa.viewSummaryRates(), true, false)
	fa.pages.AddPage(pageSumDetailName, fa.viewSumDetail(), true, false)

	// Start with a summary view
	if err := fa.app.SetRoot(fa.pages, true).Run(); err != nil {
		return err
	}
	return nil
}

func (fa *FlowApp) Stop() {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	if fa.app != nil {
		fa.app.Stop()
		fa.app = nil
	}
}
