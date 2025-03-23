package tui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type FlowApp struct {
	mu      *sync.Mutex
	app     *tview.Application
	fds     *flowdata.FlowDataStore
	fc      *flowcache.FlowCache
	stopped bool
}

func NewFlowApp(fds *flowdata.FlowDataStore, fc *flowcache.FlowCache) *FlowApp {
	return &FlowApp{&sync.Mutex{}, tview.NewApplication(), fds, fc, false}
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
			return event
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

func (fa *FlowApp) Stop() {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	if fa.app != nil && !fa.stopped {
		fa.app.Stop()
		fa.stopped = true
	}
}

func newTable() *tview.Table {
	t := tview.NewTable().SetBorders(false).SetSelectable(false, false)
	applyTheme(t)
	return t
}

var (
	bgColor       = tcell.ColorBlack
	textColor     = tcell.ColorLightGray
	borderColor   = tcell.ColorDarkSlateGray
	titleColor    = tcell.ColorWhite
	selectedStyle = tcell.StyleDefault.Foreground(titleColor).Background(tcell.ColorDarkBlue)
	tcellValStyle = tcell.StyleDefault.Background(bgColor).Foreground(textColor)
	tcellHdrStyle = tcellValStyle.Bold(true).Underline(false)
)

func applyTheme(components ...tview.Primitive) {
	for _, p := range components {
		switch c := p.(type) {
		case *tview.Flex:
			c.SetBackgroundColor(bgColor)
			c.SetTitleColor(titleColor)
			c.SetBorderColor(borderColor)
		case *tview.Table:
			c.SetBackgroundColor(bgColor)
			c.SetBorderColor(borderColor)
			c.SetTitleColor(titleColor)
			c.SetSelectedStyle(selectedStyle)
		case *tview.TextView:
			c.SetBackgroundColor(bgColor)
			c.SetTextColor(textColor)
			c.SetBorderColor(borderColor)
			c.SetTitleColor(titleColor)
		case *tview.Box:
			c.SetBackgroundColor(bgColor)
			c.SetBorderColor(borderColor)
			c.SetTitleColor(titleColor)
		}
	}
}

func (fa *FlowApp) viewSummary(selectRow int) *tview.Application {
	tableData := newTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowSumTable{fc: fa.fc}).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			sumRow, _ := tableData.GetSelection()
			if sumRow > 0 {
				sumID := tableData.GetCell(sumRow, 0).GetReference().(int)
				fa.viewSumDetail(sumID, sumRow, 0)
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
	applyTheme(flex, tableData)
	fa.app.SetRoot(flex, true)
	return fa.app
}

func (fa *FlowApp) viewSumDetail(sumID, sumRow, sumDetailRow int) *tview.Application {
	tableKeyHeader := newTable().SetBorders(true).SetSelectable(false, false).
		SetContent(&flowKeyHeaderTable{fds: fa.fds, sumID: sumID})

	dt := &flowSumDetailTable{fc: fa.fc, sumID: sumID}
	tableData := newTable().SetBorders(false).SetSelectable(true, false).
		SetContent(dt).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			sumDetailRow, _ := tableData.GetSelection()
			if sumDetailRow > 0 {
				flowID := tableData.GetCell(sumDetailRow, 0).GetReference().(int)
				fa.viewFlowDetail(sumID, flowID, sumRow, sumDetailRow)
				return nil
			}
		case tcell.KeyEscape:
			fa.viewSummary(sumRow)
			return nil
		}
		return event
	})

	tableData.SetFocusFunc(func() {
		tableData.Select(sumDetailRow, 0)
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Summary Detail")
	flex.AddItem(tableKeyHeader, 6, 1, false)
	flex.AddItem(tableData, 0, 1, true)
	applyTheme(flex, tableKeyHeader, tableData)
	fa.app.SetRoot(flex, true)
	return fa.app
}

func (fa *FlowApp) viewFlowDetail(sumID, flowID, sumRow, sumDetailRow int) *tview.Application {
	fd := fa.fds.GetFlowDetail(flowID)
	fdht := NewFlowDetailHeaderTable(fd)
	tableDetailHeader := newTable().SetBorders(true).SetSelectable(false, false).SetContent(fdht)

	viewText := fmt.Sprintf("SRC LABELS: %s\n\nDST LABELS: %s\n\nPolicy Hits Enforced:\n\n%sPolicy Hits Pending:\n\n%s",
		fd.SourceLabels, fd.DestLabels, policyHitsToString(fd.Policies.Enforced), policyHitsToString(fd.Policies.Pending))

	moreDetails := tview.NewTextView()
	moreDetails.SetText(viewText).SetBorder(true).SetTitle("More Details")

	moreDetails.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			fa.viewSumDetail(sumID, sumRow, sumDetailRow)
			return nil
		}
		return event
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Detail")
	flex.AddItem(tableDetailHeader, 6, 1, false)
	flex.AddItem(moreDetails, 0, 1, true)
	applyTheme(flex, tableDetailHeader, moreDetails)
	fa.app.SetRoot(flex, true)
	return fa.app
}

func policyHitsToString(policyHits []*flowdata.PolicyHit) string {
	var s string
	for _, ph := range policyHits {
		s += fmt.Sprintf("%s\n", policyHitToString(ph))
	}
	return s
}

func policyHitToString(ph *flowdata.PolicyHit) string {
	phs := fmt.Sprintf("Kind: %s\nName: %s\nNamespace: %s\nTier: %s\nAction: %s\nPolicyIndex: %d\nRuleIndex: %d\n",
		ph.Kind, ph.Name, ph.Namespace, ph.Tier, ph.Action, ph.PolicyIndex, ph.RuleIndex)
	if ph.Trigger != nil {
		phs += "\nTriggers:\n" + policyHitTriggerToString(ph.Trigger)
	}
	return phs
}

func policyHitTriggerToString(ph *flowdata.PolicyHit) string {
	phs := fmt.Sprintf("\tKind: %s\n\tName: %s\n\tNamespace: %s\n\tTier: %s\n\tAction: %s\n\tPolicyIndex: %d\n\tRuleIndex: %d\n",
		ph.Kind, ph.Name, ph.Namespace, ph.Tier, ph.Action, ph.PolicyIndex, ph.RuleIndex)
	if ph.Trigger != nil {
		phs += "\n\n" + policyHitTriggerToString(ph.Trigger)
	}
	return phs
}

// package main
//
// import (
// 	"github.com/gdamore/tcell/v2"
// 	"github.com/rivo/tview"
// )
//
// func main() {
// 	app := tview.NewApplication()
//
// 	// Create the main layout
// 	mainLayout := tview.NewFlex().
// 		SetDirection(tview.FlexRow).
// 		AddItem(tview.NewTextView().
// 			SetTextAlign(tview.AlignCenter).
// 			SetText("Press 'Ctrl+F' to open the form dialog"), 0, 1, false)
//
// 	// Create the form that will go in the modal
// 	form := tview.NewForm().
// 		AddInputField("Name", "", 20, nil, nil).
// 		AddInputField("Email", "", 30, nil, nil).
// 		AddPasswordField("Password", "", 20, '*', nil).
// 		AddCheckbox("Send welcome email", false, nil).
// 		AddTextArea("Notes", "", 40, 4, 0, nil).
// 		AddButton("Save", func() {
// 			// Save form data here
// 			app.QueueUpdateDraw(func() {
// 				// Close the modal
// 				pages.SwitchToPage("main")
// 			})
// 		}).
// 		AddButton("Cancel", func() {
// 			app.QueueUpdateDraw(func() {
// 				// Close the modal without saving
// 				pages.SwitchToPage("main")
// 			})
// 		})
//
// 	// Style the form
// 	form.SetBorder(true).
// 		SetTitle("User Information").
// 		SetTitleAlign(tview.AlignCenter).
// 		SetBorderColor(tcell.ColorSteelBlue)
//
// 	// Create a flex container for the modal to center the form
// 	modalFlex := tview.NewFlex().
// 		SetDirection(tview.FlexRow).
// 		AddItem(nil, 0, 1, false).
// 		AddItem(tview.NewFlex().
// 			SetDirection(tview.FlexColumn).
// 			AddItem(nil, 0, 1, false).
// 			AddItem(form, 50, 1, true). // Width of 50
// 			AddItem(nil, 0, 1, false),
// 			10, 1, true). // Height of 10
// 		AddItem(nil, 0, 1, false)
//
// 	// Create modal frame with semi-transparent background
// 	modal := tview.NewFlex().
// 		SetBackgroundColor(tcell.ColorBlack.WithAlpha(192)) // Semi-transparent black
// 	modal.AddItem(modalFlex, 0, 1, true)
//
// 	// Create pages to switch between main and modal
// 	pages := tview.NewPages().
// 		AddPage("main", mainLayout, true, true).
// 		AddPage("modal", modal, true, false) // Modal starts hidden
//
// 	// Add keyboard shortcuts
// 	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
// 		if event.Key() == tcell.KeyCtrlF {
// 			// Show the modal when Ctrl+F is pressed
// 			pages.ShowPage("modal")
// 			return nil
// 		} else if event.Key() == tcell.KeyEscape {
// 			// Hide the modal when Escape is pressed
// 			if pages.HasPage("modal") {
// 				pages.HidePage("modal")
// 				return nil
// 			}
// 		}
// 		return event
// 	})
//
// 	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
// 		panic(err)
// 	}
// }
