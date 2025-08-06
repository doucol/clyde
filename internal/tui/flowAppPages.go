package tui

import (
	"context"
	"fmt"

	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/util"
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

func (fa *FlowApp) viewHomePage(ctx context.Context) tview.Primitive {
	cc := cmdctx.CmdCtxFromContext(ctx)
	clientset := cc.Clientset()
	dyn := cc.ClientDyn()
	restConfig := cc.GetK8sConfig()
	info := util.GetClusterNetworkingInfo(ctx, clientset, dyn, restConfig)

	list := tview.NewList()
	list.SetBorder(true).SetTitle("Clyde - Main Menu")

	// Option 1: General Kubernetes Network information
	list.AddItem("General Kubernetes Network Information", "View CNI, IP ranges, and network configuration", '1', func() {
		fa.showNetworkInfoModal(info)
	})

	// Option 2: General Calico information (if Calico is installed)
	if info.CalicoInstalled {
		list.AddItem("General Calico Information", "View Calico version, operator status, and configuration", '2', func() {
			fa.showCalicoInfoModal(info)
		})
	}

	// Option 3 & 4: Flow pages (if Calico v3.30+ and whisker available)
	canShowFlows := info.CalicoInstalled && info.OperatorInstalled &&
		util.CompareVersions(info.CalicoVersion, "3.30.0") && info.WhiskerAvailable

	if canShowFlows {
		list.AddItem("Flow Sum Totals", "View flow summary totals", '3', func() {
			fa.pages.SwitchToPage(pageSummaryTotalsName)
		})

		list.AddItem("Flow Sum Rates", "View flow summary rates", '4', func() {
			fa.pages.SwitchToPage(pageSummaryRatesName)
		})
	}

	// Option 5: Show help
	list.AddItem("Help", "Show help information", 'h', func() {
		fa.showHelpDialog()
	})

	// Option 6: Exit
	list.AddItem("Exit", "Exit the application", 'q', func() {
		fa.Stop()
		cc.Cancel()
	})

	// Add installation option if needed
	canInstall := !info.OperatorInstalled || !info.CalicoInstalled || !util.CompareVersions(info.CalicoVersion, "3.30.0")
	if canInstall {
		list.AddItem("Install/Upgrade Calico Operator v3.30 (WARNING: experimental!)", "Install or upgrade Calico operator", 'i', func() {
			fa.installCalicoOperator(ctx)
		})
	}

	applyTheme(list)

	// Create a centered layout
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add flexible space above (to center vertically)
	mainFlex.AddItem(nil, 0, 1, false)

	// Add horizontal centering flex
	horizontalFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	horizontalFlex.AddItem(nil, 0, 1, false)  // Left padding
	horizontalFlex.AddItem(list, 80, 0, true) // Centered menu with fixed width
	horizontalFlex.AddItem(nil, 0, 1, false)  // Right padding

	mainFlex.AddItem(horizontalFlex, 0, 2, true) // Menu area (takes 2/3 of available space)

	// Add flexible space below (to center vertically)
	mainFlex.AddItem(nil, 0, 1, false)

	applyTheme(mainFlex)
	return mainFlex
}

func (fa *FlowApp) showNetworkInfoModal(info util.ClusterNetworkingInfo) {
	text := fmt.Sprintf(
		"[white]CNI Type: [yellow]%s\n[white]Pod CIDRs: [yellow]%s\n[white]Service CIDRs: [yellow]%s\n[white]Overlay: [yellow]%s\n[white]Encapsulation: [yellow]%s\n",
		info.CNIType,
		func() string {
			if len(info.PodCIDRs) == 0 {
				return "Not detected"
			}
			result := ""
			for i, cidr := range info.PodCIDRs {
				if i > 0 {
					result += ", "
				}
				result += cidr
			}
			return result
		}(),
		func() string {
			if len(info.ServiceCIDRs) == 0 {
				return "Not detected"
			}
			result := ""
			for i, cidr := range info.ServiceCIDRs {
				if i > 0 {
					result += ", "
				}
				result += cidr
			}
			return result
		}(),
		func() string {
			if info.Overlay == "" {
				return "Not detected"
			}
			return info.Overlay
		}(),
		func() string {
			if info.Encapsulation == "" {
				return "Not detected"
			}
			return info.Encapsulation
		}(),
	)

	if len(info.Errors) > 0 {
		text += "[red]Errors: "
		for i, err := range info.Errors {
			if i > 0 {
				text += "; "
			}
			text += err
		}
		text += "\n"
	}

	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			fa.pages.RemovePage("networkInfo")
		})
	fa.pages.AddPage("networkInfo", modal, true, true)
}

func (fa *FlowApp) showCalicoInfoModal(info util.ClusterNetworkingInfo) {
	text := fmt.Sprintf(
		"[white]Calico Installed: [yellow]%v\n[white]Calico Version: [yellow]%s\n[white]Operator Installed: [yellow]%v\n[white]Operator Version: [yellow]%s\n[white]Whisker Available: [yellow]%v\n",
		info.CalicoInstalled,
		func() string {
			if info.CalicoVersion == "" {
				return "Not detected"
			}
			return info.CalicoVersion
		}(),
		info.OperatorInstalled,
		func() string {
			if info.OperatorVersion == "" {
				return "Not detected"
			}
			return info.OperatorVersion
		}(),
		info.WhiskerAvailable,
	)

	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			fa.pages.RemovePage("calicoInfo")
		})
	fa.pages.AddPage("calicoInfo", modal, true, true)
}

func (fa *FlowApp) installCalicoOperator(ctx context.Context) {
	go func() {
		err := util.InstallCalicoOperator(ctx)
		fa.app.QueueUpdateDraw(func() {
			modal := tview.NewModal().SetText("Calico operator install: " + func() string {
				if err != nil {
					return "Failed: " + err.Error()
				}
				return "Success!"
			}()).AddButtons([]string{"OK"}).SetDoneFunc(func(_ int, _ string) {
				fa.pages.RemovePage("installResult")
			})
			fa.pages.AddPage("installResult", modal, true, true)
		})
	}()
}
