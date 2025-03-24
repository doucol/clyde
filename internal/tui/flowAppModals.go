package tui

import (
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (fa *FlowApp) filterModal() {
	const modalName = "filterModal"

	gs := global.GetState()

	actionIdx := 0
	switch gs.Filter.Action {
	case "Deny":
		actionIdx = 1
	case "Allow":
		actionIdx = 2
	}

	reporterdx := 0
	switch gs.Filter.Reporter {
	case "Src":
		reporterdx = 1
	case "Dst":
		reporterdx = 2
	}

	form := tview.NewForm()
	form.
		AddDropDown("Action", []string{"All", "Deny", "Allow"}, actionIdx, nil).
		AddDropDown("Reporter", []string{"All", "Src", "Dst"}, reporterdx, nil).
		AddInputField("Namespace", gs.Filter.Namespace, 60, nil, nil).
		AddInputField("Name", gs.Filter.Name, 60, nil, nil).
		// AddInputField("From (yyyy/mm/ddT00:00:00Z)", "", 20, nil, nil).
		// AddInputField("To (yyyy/mm/ddT00:00:00Z)", "", 20, nil, nil).
		AddButton("Save", func() {
			_, action := form.GetFormItemByLabel("Action").(*tview.DropDown).GetCurrentOption()
			if action == "All" {
				action = ""
			}
			_, reporter := form.GetFormItemByLabel("Reporter").(*tview.DropDown).GetCurrentOption()
			if reporter == "All" {
				reporter = ""
			}
			namespace := form.GetFormItemByLabel("Namespace").(*tview.InputField).GetText()
			name := form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
			// from := form.GetFormItemByLabel("From (yyyy/mm/ddT00:00:00Z)").(*tview.InputField).GetText()
			// to := form.GetFormItemByLabel("To (yyyy/mm/ddT00:00:00Z)").(*tview.InputField).GetText()
			filter := &flowdata.FilterAttributes{}
			filter.Action = action
			filter.Reporter = reporter
			filter.Namespace = namespace
			filter.Name = name
			// fa.filter.From = ""
			// fa.filter.To = ""
			gs := global.GetState()
			gs.Filter = *filter
			global.SetState(gs)
			fa.pages.RemovePage(modalName)
		}).
		AddButton("Cancel", func() {
			fa.pages.RemovePage(modalName)
		})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			fa.pages.RemovePage(modalName)
			return nil
		}
		return event
	})

	// Style the form
	form.SetBorder(true).
		SetTitle("Filter Attributes").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(borderColor)

	// Create a flex container for the modal to center the form
	modalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(form, 80, 1, true). // Width of 80
			AddItem(nil, 0, 1, false),
			30, 1, true). // Height of 30
		AddItem(nil, 0, 1, false)

	modal := tview.NewFlex()
	modal.AddItem(modalFlex, 0, 1, true)

	fa.pages.AddPage(modalName, modal, true, true)
}
