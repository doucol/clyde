package whisker

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	colTitles     = []string{"SRC NAMESPACE", "SRC NAME", "DST NAMESPACE", "DST NAME", "ACTION"}
	colTitleStyle = tcell.Style{}.Bold(true)
)

type flowTable struct {
	tview.TableContentReadOnly
}

func (ftd *flowTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return tview.NewTableCell(colTitles[column]).SetMaxWidth(1).SetExpansion(1).SetStyle(colTitleStyle)
	}
	if val, ok := fds.Get(row); ok {
		switch column {
		case 0:
			return tview.NewTableCell(val.SourceNamespace).SetMaxWidth(1).SetExpansion(1)
		case 1:
			return tview.NewTableCell(val.SourceName).SetMaxWidth(1).SetExpansion(1)
		case 2:
			return tview.NewTableCell(val.DestNamespace).SetMaxWidth(1).SetExpansion(1)
		case 3:
			return tview.NewTableCell(val.DestName).SetMaxWidth(1).SetExpansion(1)
		case 4:
			return tview.NewTableCell(val.Action).SetMaxWidth(1).SetExpansion(1)
		}
	}
	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (ftd *flowTable) GetRowCount() int {
	return fds.Count() + 1
}

func (ftd *flowTable) GetColumnCount() int {
	return len(colTitles)
}
