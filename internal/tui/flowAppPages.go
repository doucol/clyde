package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (fa *FlowApp) viewSummary() tview.Primitive {
	tableData := newTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowSumTable{fc: fa.fc}).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			fa.fas.SumRow, _ = tableData.GetSelection()
			if fa.fas.SumRow > 0 {
				fa.fas.SumID = tableData.GetCell(fa.fas.SumRow, 0).GetReference().(int)
				fa.pages.SwitchToPage(pageSumDetailName)
				return nil
			}
		}
		return event
	})

	tableData.SetFocusFunc(func() {
		tableData.Select(fa.fas.SumRow, 0)
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Summary")
	flex.AddItem(tableData, 0, 1, true)
	applyTheme(flex, tableData)
	return flex
}

func (fa *FlowApp) viewSumDetail() tview.Primitive {
	tableKeyHeader := newTable().SetBorders(true).SetSelectable(false, false).
		SetContent(&flowKeyHeaderTable{fds: fa.fds, fas: fa.fas}).SetFixed(1, 0)

	dt := &flowSumDetailTable{fc: fa.fc, fas: fa.fas}
	tableData := newTable().SetBorders(false).SetSelectable(true, false).
		SetContent(dt).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			fa.fas.FlowRow, _ = tableData.GetSelection()
			if fa.fas.FlowRow > 0 {
				fa.fas.FlowID = tableData.GetCell(fa.fas.FlowRow, 0).GetReference().(int)
				fa.pages.AddAndSwitchToPage(pageFlowDetailName, fa.viewFlowDetail(), true)
				return nil
			}
		case tcell.KeyEscape:
			fa.pages.SwitchToPage(pageSummaryName)
			return nil
		}
		return event
	})

	tableData.SetFocusFunc(func() {
		tableData.Select(fa.fas.FlowRow, 0)
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Summary Detail")
	flex.AddItem(tableKeyHeader, 6, 1, false)
	flex.AddItem(tableData, 0, 1, true)
	applyTheme(flex, tableKeyHeader, tableData)
	return flex
}

func (fa *FlowApp) viewFlowDetail() tview.Primitive {
	tableDetailHeader := newTable().SetBorders(true).SetSelectable(false, false)
	tableDetailHeader.SetContent(NewFlowDetailHeaderTable(fa.fds, fa.fas))

	moreDetails := tview.NewTextView()
	moreDetails.SetBorder(true).SetTitle("More Details")

	fd := fa.fds.GetFlowDetail(fa.fas.FlowID)
	viewText := fmt.Sprintf("SRC LABELS: %s\n\nDST LABELS: %s\n\nPolicy Hits Enforced:\n\n%sPolicy Hits Pending:\n\n%s",
		fd.SourceLabels, fd.DestLabels, policyHitsToString(fd.Policies.Enforced), policyHitsToString(fd.Policies.Pending))
	moreDetails.SetText(viewText)

	moreDetails.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			fa.pages.SwitchToPage(pageSumDetailName)
			fa.pages.RemovePage(pageFlowDetailName)
			return nil
		}
		return event
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Detail")
	flex.AddItem(tableDetailHeader, 6, 1, false)
	flex.AddItem(moreDetails, 0, 1, true)
	applyTheme(flex, tableDetailHeader, moreDetails)
	return flex
}
