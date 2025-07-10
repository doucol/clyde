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
	table.SetDoneFunc(func(key tcell.Key) {
		fa.pages.RemovePage(helpDialogName)
	})
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Close on ESC or '?' again
		if event.Key() == tcell.KeyEscape || (event.Key() == tcell.KeyRune && event.Rune() == '?') {
			fa.pages.RemovePage(helpDialogName)
			return nil
		}
		return event
	})
	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(table, 0, 1, true)
	fa.pages.AddPage(helpDialogName, modal, true, true)
}
