package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

func (fa *FlowApp) viewSummary() tview.Primitive {
	tbl := newFlowSumTable(fa.fc)
	tableData := newTable().SetBorders(false).SetSelectable(true, false).
		SetContent(tbl).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			if fa.fas.sumRow > 0 {
				logrus.Debugf("flow state: : %+v", fa.fas)
				fa.pages.SwitchToPage(pageSumDetailName)
				return nil
			}
		}
		return event
	})

	tableData.SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		if tableData.GetRowCount() > 1 {
			row, _ := tableData.GetSelection()
			fa.fas.setSum(tableData.GetCell(row, 0).GetReference().(int), row)
		} else {
			fa.fas.setSum(0, 0)
		}
		return tableData.GetInnerRect()
	})

	tableData.SetSelectionChangedFunc(func(row, column int) {
		fa.fas.setSum(tableData.GetCell(row, 0).GetReference().(int), row)
	})
	tableData.SetSelectedFunc(func(row, column int) {
		fa.fas.setSum(tableData.GetCell(row, 0).GetReference().(int), row)
	})

	tableData.SetFocusFunc(func() {
		fa.fas.lastHomePage = pageSummaryTotalsName
		if fa.fas.sumRow == 0 && tableData.GetRowCount() > 1 {
			tableData.Select(1, 0)
		} else {
			tableData.Select(fa.fas.sumRow, 0)
		}
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Summary Totals")
	flex.AddItem(tableData, 0, 1, true)
	applyTheme(flex, tableData)
	return flex
}

func (fa *FlowApp) viewSummaryRates() tview.Primitive {
	tbl := newFlowSumRateTable(fa.fc)
	tableData := newTable().SetBorders(false).SetSelectable(true, false).
		SetContent(tbl).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			if fa.fas.rateRow > 0 {
				logrus.Debugf("flow state: : %+v", fa.fas)
				fa.pages.SwitchToPage(pageSumDetailName)
				return nil
			}
		}
		return event
	})

	tableData.SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		if tableData.GetRowCount() > 1 {
			row, _ := tableData.GetSelection()
			fa.fas.setRate(tableData.GetCell(row, 0).GetReference().(int), row)
		} else {
			fa.fas.setRate(0, 0)
		}
		return tableData.GetInnerRect()
	})

	tableData.SetSelectionChangedFunc(func(row, column int) {
		fa.fas.setRate(tableData.GetCell(row, 0).GetReference().(int), row)
	})
	tableData.SetSelectedFunc(func(row, column int) {
		fa.fas.setRate(tableData.GetCell(row, 0).GetReference().(int), row)
	})

	tableData.SetFocusFunc(func() {
		fa.fas.lastHomePage = pageSummaryRatesName
		if fa.fas.rateRow == 0 && tableData.GetRowCount() > 1 {
			tableData.Select(1, 0)
		} else {
			tableData.Select(fa.fas.rateRow, 0)
		}
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Summary Rates")
	flex.AddItem(tableData, 0, 1, true)
	applyTheme(flex, tableData)
	return flex
}

func (fa *FlowApp) viewSumDetail() tview.Primitive {
	tableKeyHeader := newTable().SetBorders(true).SetSelectable(false, false).
		SetContent(newFlowKeyHeaderTable(fa.fds, fa.fas)).SetFixed(1, 0)

	dt := newFlowSumDetailTable(fa.fc, fa.fas)
	tableData := newTable().SetBorders(false).SetSelectable(true, false).
		SetContent(dt).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if fa.fas.flowRow > 0 {
				logrus.Debugf("flow state: : %+v", fa.fas)
				fa.pages.AddAndSwitchToPage(pageFlowDetailName, fa.viewFlowDetail(), true)
				return nil
			}
		case tcell.KeyEscape:
			fa.pages.SwitchToPage(fa.fas.lastHomePage)
			return nil
		}
		return event
	})

	tableData.SetSelectionChangedFunc(func(row, column int) {
		fa.fas.setFlow(tableData.GetCell(row, 0).GetReference().(int), row)
	})
	tableData.SetSelectedFunc(func(row, column int) {
		fa.fas.setFlow(tableData.GetCell(row, 0).GetReference().(int), row)
	})

	tableData.SetFocusFunc(func() {
		if fa.fas.flowRow == 0 && tableData.GetRowCount() > 1 {
			tableData.Select(1, 0)
		} else {
			tableData.Select(fa.fas.flowRow, 0)
		}
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
	tableDetailHeader.SetContent(newFlowDetailHeaderTable(fa.fds, fa.fas))

	moreDetails := tview.NewTextView()
	moreDetails.SetBorder(true).SetTitle("More Details")

	fd := fa.fds.GetFlowDetail(fa.fas.flowID)
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
