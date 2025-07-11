package tui

import (
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
)

const helpDialogName = "helpDialog"

var helpEntries = []struct{
	Key, Description string
}{
	{"q or Ctrl+C", "Quit the application"},
	{"r", "Switch to Rates view (from Totals)"},
	{"t", "Switch to Totals view (from Rates)"},
	{"p", "Sort by Source Packet Rate (Rates view)"},
	{"P", "Sort by Dest Packet Rate (Rates view)"},
	{"b", "Sort by Source Byte Rate (Rates view)"},
	{"B", "Sort by Dest Byte Rate (Rates view)"},
	{"n", "Sort by Key (current page)"},
	{"/", "Open filter dialog"},
	{"?", "Show this help dialog"},
}

func (fa *FlowApp) showHelpDialog() {
	table := tview.NewTable().SetBorders(false)
	table.SetTitle("Help - Key Commands").SetBorder(true)
	table.SetCell(0, 0, tview.NewTableCell("Key").SetSelectable(false).SetAttributes(tcell.AttrBold))
	table.SetCell(0, 1, tview.NewTableCell("Description").SetSelectable(false).SetAttributes(tcell.AttrBold))
	for i, entry := range helpEntries {
		table.SetCell(i+1, 0, tview.NewTableCell(entry.Key))
		table.SetCell(i+1, 1, tview.NewTableCell(entry.Description))
	}

table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || (event.Key() == tcell.KeyRune && event.Rune() == '?') {
			fa.pages.RemovePage(helpDialogName)
			return nil
		}
		return event
	})

table.SetDoneFunc(func(key tcell.Key) {
		fa.pages.RemovePage(helpDialogName)
	})

	// Center the table in a modal-like Flex
	modalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(table, 60, 1, true).
			AddItem(nil, 0, 1, false),
			len(helpEntries)+4, 1, true).
		AddItem(nil, 0, 1, false)

	fa.pages.AddPage(helpDialogName, modalFlex, true, true)
}
