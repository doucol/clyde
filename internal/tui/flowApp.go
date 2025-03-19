package tui

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type FlowApp struct {
	mu      *sync.Mutex
	app     *tview.Application
	fds     *flowdata.FlowDataStore
	stopped bool
}

func NewFlowApp(fds *flowdata.FlowDataStore) *FlowApp {
	return &FlowApp{&sync.Mutex{}, tview.NewApplication(), fds, false}
}

func (fa *FlowApp) Run(ctx context.Context) error {
	defer fa.Stop()
	cmdctx := cmdContext.CmdContextFromContext(ctx)

	// Set up an input capture to shutdown the app when the user presses Ctrl-C
	fa.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			fa.Stop()
			cmdctx.Cancel()
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

	fa.setTheme()

	// Start with a summary view
	if err := fa.viewSummary(0).Run(); err != nil {
		return err
	}
	return nil
}

func (fa *FlowApp) setTheme() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.BorderColor = tcell.ColorBlue
	tview.Styles.TitleColor = tcell.ColorWhite
	tview.Styles.GraphicsColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorWhite
	tview.Styles.TertiaryTextColor = tcell.ColorWhite
	tview.Styles.InverseTextColor = tcell.ColorWhite
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorWhite
}

func concatCells(td *tview.Table, row int, sep string, cols ...int) string {
	s := []string{}
	for i := range cols {
		s = append(s, strings.TrimSpace(td.GetCell(row, cols[i]).Text))
	}
	return strings.Join(s, sep)
}

func (fa *FlowApp) Stop() {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	if fa.app != nil && !fa.stopped {
		fa.app.Stop()
		fa.stopped = true
	}
}

func (fa *FlowApp) viewSummary(selectRow int) *tview.Application {
	tableData := tview.NewTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowSumTable{fds: fa.fds}).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, _ := tableData.GetSelection()
			if row > 0 {
				key := concatCells(tableData, row, "|",
					SUMCOL_SRC_NAMESPACE, SUMCOL_SRC_NAME, SUMCOL_DST_NAMESPACE, SUMCOL_DST_NAME, SUMCOL_PROTO, SUMCOL_PORT)
				fa.viewDetail(row, key)
				return nil
			}
		}
		return event
	})

	tableData.SetFocusFunc(func() {
		tableData.Select(selectRow, 0)
	})

	fa.setTheme()
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Summary")
	flex.AddItem(tableData, 0, 1, true)
	fa.app.SetRoot(flex, true)
	return fa.app
}

func (fa *FlowApp) viewDetail(row int, key string) *tview.Application {
	tableDataHeader := tview.NewTable().SetBorders(true).SetSelectable(true, false).
		SetContent(&flowKeyHeaderTable{fds: fa.fds, key: key})

	tableData := tview.NewTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowSumDetailTable{fds: fa.fds, key: key}).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			fa.viewSummary(row)
			return nil
		}
		return event
	})

	fa.setTheme()
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Detail")
	flex.AddItem(tableDataHeader, 6, 1, false)
	flex.AddItem(tableData, 0, 1, true)
	fa.app.SetRoot(flex, true)
	return fa.app
}
