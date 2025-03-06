package whisker

import (
	"context"
	"strings"
	"time"

	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/flowdata"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type FlowApp struct {
	ctx context.Context
	app *tview.Application
	fds *flowdata.FlowDataStore
}

func NewFlowApp(ctx context.Context, fds *flowdata.FlowDataStore) *FlowApp {
	return &FlowApp{ctx, tview.NewApplication(), fds}
}

func (fa *FlowApp) Run() error {
	cc := cmdContext.CmdContextFromContext(fa.ctx)

	// Set up an input capture to shutdown the app when the user presses Ctrl-C
	fa.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			go cc.Cancel()
			go fa.app.Stop()
			return nil
		}
		return event
	})

	// Go update screen periodically until we're shutdown
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-fa.ctx.Done():
				return
			case <-ticker.C:
				fa.app.Draw()
			}
		}
	}()

	// Start with a summary view
	if err := fa.ViewSummary(0).Run(); err != nil {
		return err
	}
	return nil
}

func concatCells(td *tview.Table, row int, sep string, cols ...int) string {
	s := []string{}
	for i := range cols {
		s = append(s, strings.TrimSpace(td.GetCell(row, cols[i]).Text))
	}
	return strings.Join(s, sep)
}

func (fa *FlowApp) ViewSummary(selectRow int) *tview.Application {
	tableData := tview.NewTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowSumTable{fds: fa.fds}).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, _ := tableData.GetSelection()
			if row > 0 {
				key := concatCells(tableData, row, "|",
					SUMCOL_SRC_NAMESPACE, SUMCOL_SRC_NAME, SUMCOL_DST_NAMESPACE, SUMCOL_DST_NAME, SUMCOL_PROTO, SUMCOL_PORT)
				fa.ViewDetail(row, key)
				return nil
			}
		}
		return event
	})

	tableData.SetFocusFunc(func() {
		tableData.Select(selectRow, 0)
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Summary")
	flex.AddItem(tableData, 0, 1, true)
	fa.app.SetRoot(flex, true)
	return fa.app
}

func (fa *FlowApp) ViewDetail(row int, key string) *tview.Application {
	tableDataHeader := tview.NewTable().SetBorders(true).SetSelectable(true, false).
		SetContent(&flowDetailTableHeader{fds: fa.fds, key: key})

	tableData := tview.NewTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowDetailTable{fds: fa.fds, key: key}).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			fa.ViewSummary(row)
			return nil
		}
		return event
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Detail")
	flex.AddItem(tableDataHeader, 6, 1, false)
	flex.AddItem(tableData, 0, 1, true)
	fa.app.SetRoot(flex, true)
	return fa.app
}
