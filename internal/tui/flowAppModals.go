package tui

import (
	"fmt"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

func (fa *FlowApp) filterModal() {
	const modalName = "filterModal"

	filter := global.GetFilter()

	actionLabel := "Action:"
	reporterLabel := "Reporter:"
	namespaceLabel := "Namespace:"
	nameLabel := "Name:"

	labelUnsel := func(l string) string {
		return fmt.Sprintf("%-12s", l)
	}
	labelSel := func(l string) string {
		return fmt.Sprintf("* %-12s", l)
	}

	actionOptions := []string{"All", "Deny", "Allow"}
	reporterOptions := []string{"All", "Src", "Dst"}

	actionIdx := 0
	switch filter.Action {
	case "Deny":
		actionIdx = 1
	case "Allow":
		actionIdx = 2
	}

	reporterIdx := 0
	switch filter.Reporter {
	case "Src":
		reporterIdx = 1
	case "Dst":
		reporterIdx = 2
	}

	actionDropDown := tview.NewDropDown()
	actionDropDown.SetLabel(labelUnsel(actionLabel))
	actionDropDown.SetFocusFunc(func() {
		actionDropDown.SetLabel(labelSel(actionLabel))
	}).SetBlurFunc(func() {
		actionDropDown.SetLabel(labelUnsel(actionLabel))
	})
	actionDropDown.SetOptions(actionOptions, func(opt string, idx int) {
		filter.Action = ""
		if idx > 0 {
			filter.Action = opt
		}
	})
	actionDropDown.SetCurrentOption(actionIdx)

	reporterDropDown := tview.NewDropDown()
	reporterDropDown.SetLabel(labelUnsel(reporterLabel))
	reporterDropDown.SetFocusFunc(func() {
		reporterDropDown.SetLabel(labelSel(reporterLabel))
	}).SetBlurFunc(func() {
		reporterDropDown.SetLabel(labelUnsel(reporterLabel))
	})
	reporterDropDown.SetOptions(reporterOptions, func(opt string, idx int) {
		filter.Reporter = ""
		if idx > 0 {
			filter.Reporter = opt
		}
	})
	reporterDropDown.SetCurrentOption(reporterIdx)

	namespaceInputField := tview.NewInputField()
	namespaceInputField.SetLabel(labelUnsel(namespaceLabel))
	namespaceInputField.SetFocusFunc(func() {
		namespaceInputField.SetLabel(labelSel(namespaceLabel))
	}).SetBlurFunc(func() {
		namespaceInputField.SetLabel(labelUnsel(namespaceLabel))
	})
	namespaceInputField.SetText(filter.Namespace)
	namespaceInputField.SetFieldWidth(60)
	namespaceInputField.SetChangedFunc(func(text string) {
		filter.Namespace = text
	})

	nameInputField := tview.NewInputField()
	nameInputField.SetLabel(labelUnsel(nameLabel))
	nameInputField.SetFocusFunc(func() {
		nameInputField.SetLabel(labelSel(nameLabel))
	}).SetBlurFunc(func() {
		nameInputField.SetLabel(labelUnsel(nameLabel))
	})
	nameInputField.SetText(filter.Name)
	nameInputField.SetFieldWidth(60)
	nameInputField.SetChangedFunc(func(text string) {
		filter.Name = text
	})

	form := tview.NewForm()
	form.AddFormItem(actionDropDown)
	form.AddFormItem(reporterDropDown)
	form.AddFormItem(namespaceInputField)
	form.AddFormItem(nameInputField)
	form.AddButton("Save", func() {
		global.SetFilter(filter)
		logrus.Debugf("Filter: %+v", filter)
		fa.pages.RemovePage(modalName)
	})
	form.AddButton("Cancel", func() {
		fa.pages.RemovePage(modalName)
	})
	form.AddButton("Clear", func() {
		global.SetFilter(flowdata.FilterAttributes{})
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
			20, 1, true). // Height of 20
		AddItem(nil, 0, 1, false)

	modal := tview.NewFlex()
	modal.AddItem(modalFlex, 0, 1, true)

	prims := []tview.Primitive{form, actionDropDown, reporterDropDown, namespaceInputField, nameInputField}
	applyTheme(prims...)
	fa.pages.AddPage(modalName, modal, true, true)
}
