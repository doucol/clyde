package tui

import (
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (fa *FlowApp) filterModal() {
	const modalName = "filterModal"

	filter := global.GetFilter()

	actionIdx := 0
	switch filter.Action {
	case "Deny":
		actionIdx = 1
	case "Allow":
		actionIdx = 2
	}

	reporterdx := 0
	switch filter.Reporter {
	case "Src":
		reporterdx = 1
	case "Dst":
		reporterdx = 2
	}

	form := tview.NewForm()
	form.
		AddDropDown("Action", []string{"All", "Deny", "Allow"}, actionIdx, nil).
		AddDropDown("Reporter", []string{"All", "Src", "Dst"}, reporterdx, nil).
		AddInputField("Namespace", filter.Namespace, 60, nil, nil).
		AddInputField("Name", filter.Name, 60, nil, nil).
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
			filter := flowdata.FilterAttributes{}
			filter.Action = action
			filter.Reporter = reporter
			filter.Namespace = namespace
			filter.Name = name
			// fa.filter.From = ""
			// fa.filter.To = ""
			global.SetFilter(filter)
			fa.pages.RemovePage(modalName)
		}).
		AddButton("Cancel", func() {
			fa.pages.RemovePage(modalName)
		})

		// form.SetFieldStyle(tcell.StyleDefault.Attributes(tcell.AttrMask))

		// Change label color on focus
	// form.SetFormItemStyles(tview.Styles{
	// 	FieldStyle: tview.FieldStyle{
	// 		LabelActivated: tcell.ColorRed,
	// 	},
	// })

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			fa.pages.RemovePage(modalName)
			return nil
		}

		// focusIdx, _ := form.GetFocusedItemIndex()
		// for idx := 0; idx < form.GetFormItemCount(); idx += 1 {
		// 	border := false
		// 	style := bgColor
		// 	if idx == focusIdx {
		// 		border = true
		// 		color = textColor
		// 	}
		// 	item := form.GetFormItem(idx)
		// 	switch fi := item.(type) {
		// 	case *tview.DropDown:
		// 		logrus.Debugf("In dropdown label setter: %d, %d", idx, focusIdx)
		// 		fi.SetBorder(border)
		//       fi.SetBorderStyle(style tcell.Style)
		// 		fi.SetBorderColor(color)
		// 	case *tview.InputField:
		// 		logrus.Debugf("In InputField label setter: %d, %d", idx, focusIdx)
		// 		fi.SetBorder(border)
		// 		fi.SetBorderColor(color)
		// 	}
		// }
		//
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

	prims := []tview.Primitive{
		form,
		form.GetFormItemByLabel("Action"),
		form.GetFormItemByLabel("Reporter"),
		form.GetFormItemByLabel("Namespace"),
		form.GetFormItemByLabel("Name"),
	}

	applyTheme(prims...)
	fa.pages.AddPage(modalName, modal, true, true)
}

// package main
//
// import (
// 	"github.com/rivo/tview"
// )
//
// func main() {
// 	app := tview.NewApplication()
//
// 	// Create the form
// 	form := tview.NewForm()
//
// 	// Add form items
// 	inputField := tview.NewInputField().SetLabel("Name:").SetFieldWidth(20)
// 	form.AddFormItem(inputField)
//
// 	dropDown := tview.NewDropDown().
// 		SetLabel("Title:").
// 		SetOptions([]string{"Mr.", "Ms.", "Dr.", "Prof."}, nil).
// 		SetFieldWidth(10)
// 	form.AddFormItem(dropDown)
//
// 	checkBox := tview.NewCheckbox().SetLabel("Subscribe:").SetChecked(true)
// 	form.AddFormItem(checkBox)
//
// 	textArea := tview.NewTextArea().
// 		SetLabel("Bio:").
// 		SetText("").
// 		SetFieldHeight(3).
// 		SetFieldWidth(30)
// 	form.AddFormItem(textArea)
//
// 	// Add buttons (these don't get the special focus treatment)
// 	form.AddButton("Save", func() {
// 		// Handle save
// 		app.Stop()
// 	})
// 	form.AddButton("Cancel", func() {
// 		app.Stop()
// 	})
//
// 	// Keep track of the currently focused item.  We need this for redrawing
// 	// when focus changes.
// 	var currentlyFocusedItem tview.FormItem
//
// 	// setFocusedItem updates the border and redraws the necessary parts.
// 	setFocusedItem := func(item tview.FormItem) {
// 		if currentlyFocusedItem != nil {
// 			// Reset attributes of the previously focused item.
// 			currentlyFocusedItem.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor) // Use default background
// 			currentlyFocusedItem.SetLabelColor(tview.Styles.SecondaryTextColor)              //Use default label color
//
// 		}
//
// 		if item != nil {
// 			// Set new focus styles
// 			item.SetFieldBackgroundColor(tview.ColorBlue)        // Highlight the background
// 			item.SetLabelColor(tview.ColorYellow)             // Make label stand out
// 		}
//
// 		currentlyFocusedItem = item // Store the newly focused item
//
// 		app.Draw() // VERY IMPORTANT: Force a redraw to apply changes!
// 	}
//
//
// 	// Set initial focus.  This is CRUCIAL.
// 	setFocusedItem(inputField)
//
// 	// Capture focus changes on the form. This is the core logic.
// 	form.SetFocusFunc(func(primitive tview.Primitive) {
// 		if formItem, ok := primitive.(tview.FormItem); ok {
// 			// It's a FormItem, so we can highlight it.
// 			setFocusedItem(formItem)
// 		} else {
// 			// No FormItem is focused (e.g., a button is focused).
// 			setFocusedItem(nil)
// 		}
// 	})
//
//
//
// 	if err := app.SetRoot(form, true).Run(); err != nil {
// 		panic(err)
// 	}
// }
