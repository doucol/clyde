package tui

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

func (fa *FlowApp) filterChange(filter flowdata.FilterAttributes) {
	f := global.GetFilter()
	if f == (filter) {
		return
	}
	global.SetFilter(filter)
	fa.fas.reset()
	fa.pages.SwitchToPage(fa.fas.lastHomePage)
	logrus.Debugf("Filter: %+v", filter)
}

const modalFilterDialogName = "modalFilterDialog"

func (fa *FlowApp) filterModal() {
	const actionLabel = "Action:"
	const portLabel = "Port:"
	const namespaceLabel = "Namespace:"
	const nameLabel = "Name:"
	const labelLabel = "Label:"
	const dateFromLabel = "Date From:"
	const dateToLabel = "Date To:"
	filter := global.GetFilter()

	actionOptions := []string{"All", "Deny", "Allow"}

	actionIdx := 0
	switch filter.Action {
	case "Deny":
		actionIdx = 1
	case "Allow":
		actionIdx = 2
	}

	actionDropDown := tview.NewDropDown()
	actionDropDown.SetLabel(actionLabel)
	actionDropDown.SetOptions(actionOptions, func(opt string, idx int) {
		filter.Action = ""
		if idx > 0 {
			filter.Action = opt
		}
	})
	actionDropDown.SetCurrentOption(actionIdx)

	portInputField := tview.NewInputField()
	portInputField.SetLabel(portLabel)
	if filter.Port > 0 {
		portInputField.SetText(strconv.Itoa(filter.Port))
	}
	portInputField.SetFieldWidth(6)
	portInputField.SetAcceptanceFunc(tview.InputFieldInteger)
	portInputField.SetChangedFunc(func(text string) {
		filter.Port, _ = strconv.Atoi(text)
	})

	namespaceInputField := tview.NewInputField()
	namespaceInputField.SetLabel(namespaceLabel)
	namespaceInputField.SetText(filter.Namespace)
	namespaceInputField.SetAcceptanceFunc(tview.InputFieldMaxLength(60))
	namespaceInputField.SetFieldWidth(60)
	namespaceInputField.SetChangedFunc(func(text string) {
		filter.Namespace = text
	})

	nameInputField := tview.NewInputField()
	nameInputField.SetLabel(nameLabel)
	nameInputField.SetText(filter.Name)
	nameInputField.SetAcceptanceFunc(tview.InputFieldMaxLength(60))
	nameInputField.SetFieldWidth(60)
	nameInputField.SetChangedFunc(func(text string) {
		filter.Name = text
	})

	labelInputField := tview.NewInputField()
	labelInputField.SetLabel(labelLabel)
	labelInputField.SetText(filter.Label)
	labelInputField.SetAcceptanceFunc(tview.InputFieldMaxLength(60))
	labelInputField.SetFieldWidth(60)
	labelInputField.SetChangedFunc(func(text string) {
		filter.Label = text
	})

	dateFromInputField := tview.NewInputField()
	dateFromInputField.SetLabel(dateFromLabel)
	dateFromInputField.SetText(tf(filter.DateFrom))
	dateFromInputField.SetAcceptanceFunc(dateAcceptanceFunc)
	dateFromInputField.SetFieldWidth(20)
	dateFromInputField.SetChangedFunc(func(text string) {
		filter.DateFrom = setDateChanged(text)
	})

	dateToInputField := tview.NewInputField()
	dateToInputField.SetLabel(dateToLabel)
	dateToInputField.SetText(tf(filter.DateTo))
	dateToInputField.SetAcceptanceFunc(dateAcceptanceFunc)
	dateToInputField.SetFieldWidth(20)
	dateToInputField.SetChangedFunc(func(text string) {
		filter.DateTo = setDateChanged(text)
	})

	form := tview.NewForm()
	form.AddFormItem(actionDropDown)
	form.AddFormItem(portInputField)
	form.AddFormItem(namespaceInputField)
	form.AddFormItem(nameInputField)
	form.AddFormItem(labelInputField)
	form.AddFormItem(dateFromInputField)
	form.AddFormItem(dateToInputField)
	form.AddButton("Save", func() {
		fa.filterChange(filter)
		fa.pages.RemovePage(modalFilterDialogName)
	})
	form.AddButton("Cancel", func() {
		fa.pages.RemovePage(modalFilterDialogName)
	})
	form.AddButton("Clear", func() {
		fa.filterChange(flowdata.FilterAttributes{})
		fa.pages.RemovePage(modalFilterDialogName)
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			fa.pages.RemovePage(modalFilterDialogName)
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			if actionDropDown.HasFocus() {
				return event
			}
			for i := 0; i < form.GetButtonCount(); i++ {
				if form.GetButton(i).HasFocus() {
					return event
				}
			}
			fa.filterChange(filter)
			fa.pages.RemovePage(modalFilterDialogName)
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

	applyTheme(form, actionDropDown, portInputField, namespaceInputField, nameInputField, labelInputField, dateFromInputField, dateToInputField)
	fa.pages.AddPage(modalFilterDialogName, modal, true, true)
}

var dateRegex = regexp.MustCompile(`^\d{0,4}?-\d{0,2}?-\d{0,2}?T\d{0,2}?:\d{0,2}?:\d{0,2}Z$`)

func dateAcceptanceFunc(text string, ch rune) bool {
	return dateRegex.Match([]byte(strings.TrimSpace(text)))
}

func setDateChanged(text string) time.Time {
	if text != "" {
		if gt, err := time.Parse(time.RFC3339, text); err == nil {
			return gt
		}
	}
	return time.Time{}
}
