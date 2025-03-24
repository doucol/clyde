package tui

import (
	"context"
	"sync"
	"time"

	"github.com/doucol/clyde/internal/cmdContext"
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
	SumID   int
	SumRow  int
	FlowID  int
	FlowRow int
}

type FlowApp struct {
	mu      *sync.Mutex
	app     *tview.Application
	fds     *flowdata.FlowDataStore
	fc      *flowcache.FlowCache
	fas     *flowAppState
	stopped bool
	pages   *tview.Pages
}

func NewFlowApp(fds *flowdata.FlowDataStore, fc *flowcache.FlowCache) *FlowApp {
	setTheme()
	pages := tview.NewPages()
	pages.SetBackgroundColor(bgColor).SetBorderColor(borderColor).SetTitleColor(titleColor)
	pages.SetBorderStyle(tcell.StyleDefault.Foreground(borderColor).Background(bgColor))
	fas := &flowAppState{}
	return &FlowApp{&sync.Mutex{}, tview.NewApplication(), fds, fc, fas, false, pages}
}

func (fa *FlowApp) Run(ctx context.Context) error {
	defer fa.Stop()
	cmdctx := cmdContext.CmdContextFromContext(ctx)

	// Set up an input capture to shutdown the app when the user presses Ctrl-C
	fa.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC || (event.Key() == tcell.KeyRune && event.Rune() == 'q') {
			fa.Stop()
			cmdctx.Cancel()
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
	if fa.app != nil && !fa.stopped {
		fa.app.Stop()
		fa.stopped = true
	}
}
