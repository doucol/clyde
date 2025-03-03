package whisker

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	colTitles     = []string{"SRC NAMESPACE", "SRC NAME", "DST NAMESPACE", "DST NAME", "PROTO", "PORT", "PACK IN", "PACK OUT", "BYTE IN", "BYTE OUT", "ACTION"}
	colTitleStyle = tcell.Style{}.Bold(true)
)

type flowTable struct {
	tview.TableContentReadOnly
}

func uintos(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func intos(v int64) string {
	return strconv.FormatInt(v, 10)
}

func (ft *flowTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return tview.NewTableCell(colTitles[column]).SetMaxWidth(1).SetExpansion(1).SetStyle(colTitleStyle)
	}
	if fa, ok := fds.GetFlowSum(row); ok {
		switch column {
		case 0:
			return tview.NewTableCell(fa.SourceNamespace).SetMaxWidth(1).SetExpansion(1)
		case 1:
			return tview.NewTableCell(fa.SourceName).SetMaxWidth(1).SetExpansion(2)
		case 2:
			return tview.NewTableCell(fa.DestNamespace).SetMaxWidth(1).SetExpansion(1)
		case 3:
			return tview.NewTableCell(fa.DestName).SetMaxWidth(1).SetExpansion(2)
		case 4:
			return tview.NewTableCell(fa.Protocol).SetMaxWidth(1).SetExpansion(0)
		case 5:
			return tview.NewTableCell(intos(fa.DestPort)).SetMaxWidth(1).SetExpansion(0)
		case 6:
			return tview.NewTableCell(uintos(fa.PacketsIn)).SetMaxWidth(1).SetExpansion(1)
		case 7:
			return tview.NewTableCell(uintos(fa.PacketsOut)).SetMaxWidth(1).SetExpansion(1)
		case 8:
			return tview.NewTableCell(uintos(fa.BytesIn)).SetMaxWidth(1).SetExpansion(1)
		case 9:
			return tview.NewTableCell(uintos(fa.BytesOut)).SetMaxWidth(1).SetExpansion(1)
		case 10:
			return tview.NewTableCell(fa.Action).SetMaxWidth(1).SetExpansion(0)
		}
	}
	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (ft *flowTable) GetRowCount() int {
	return fds.GetFlowSumCount() + 1
}

func (ft *flowTable) GetColumnCount() int {
	return len(colTitles)
}
