package tui

import (
	"context"
	"sync"
	"time"

	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

const (
	pageSummaryName    = "summary"
	pageSumDetailName  = "sumDetail"
	pageFlowDetailName = "flowDetail"
)

type flowAppState struct {
	sumID   int
	sumRow  int
	flowID  int
	flowRow int
}

func (fas *flowAppState) reset() {
	fas.sumID = 0
	fas.sumRow = 0
	fas.flowID = 0
	fas.flowRow = 0
}

func (fas *flowAppState) setSum(id, row int) {
	fas.sumID = id
	fas.sumRow = row
	fas.flowID = 0
	fas.flowRow = 0
}

func (fas *flowAppState) setFlow(id, row int) {
	fas.flowID = id
	fas.flowRow = row
}

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

func (fa *FlowApp) Run(ctx context.Context) error {
	defer fa.Stop()
	cc := cmdctx.CmdCtxFromContext(ctx)

	// Set up an input capture to shutdown the app when the user presses Ctrl-C
	fa.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC || (event.Key() == tcell.KeyRune && event.Rune() == 'q') {
			fa.Stop()
			cc.Cancel()
			return nil
		}
		if event.Key() == tcell.KeyRune && event.Rune() == '/' {
			fa.filterModal()
			return nil
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

	fa.pages.AddPage(pageSummaryName, fa.viewSummary(), true, true)
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
