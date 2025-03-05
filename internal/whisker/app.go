package whisker

import (
	"context"
	"fmt"
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
	fa.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			go cc.Cancel()
			go fa.app.Stop()
			return nil
		}
		return event
	})

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

	if err := fa.ViewSummary().Run(); err != nil {
		return err
	}
	return nil
}

func (fa *FlowApp) ViewSummary() *tview.Application {
	tableData := tview.NewTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowSumTable{fds: fa.fds}).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, _ := tableData.GetSelection()
			if row > 0 {
				sns := tableData.GetCell(row, 0).Text
				sn := tableData.GetCell(row, 1).Text
				dns := tableData.GetCell(row, 2).Text
				dn := tableData.GetCell(row, 3).Text
				proto := tableData.GetCell(row, 4).Text
				port := tableData.GetCell(row, 5).Text
				key := fmt.Sprintf("%s|%s|%s|%s|%s|%s", sns, sn, dns, dn, proto, port)
				fa.ViewDetail(key)
				return nil
			}
		}
		return event
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Summary")
	flex.AddItem(tableData, 0, 1, true)
	fa.app.SetRoot(flex, true)
	return fa.app
}

func (fa *FlowApp) ViewDetail(key string) *tview.Application {
	tableDataHeader := tview.NewTable().SetBorders(true).SetSelectable(true, false).
		SetContent(&flowDetailTableHeader{fds: fa.fds, key: key})

	tableData := tview.NewTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowDetailTable{fds: fa.fds, key: key}).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			fa.ViewSummary()
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
