package tui

import (
	"fmt"
	"time"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/rivo/tview"
)

// FlowDetailTable is a table for displaying flow details.
type flowSumDetailTable struct {
	tview.TableContentReadOnly
	fds *flowdata.FlowDataStore
	key string
}

func (fdt *flowSumDetailTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return hdrCell(dtlCols[column], 1, 1)
	}
	if fd, ok := fdt.fds.GetFlowDetail(fdt.key, row); ok {
		switch column {
		case DTLCOL_START_TIME:
			return valCell(fd.StartTime.Format(time.RFC3339), 1, 0)
		case DTLCOL_END_TIME:
			return valCell(fd.EndTime.Format(time.RFC3339), 1, 0)
		case DTLCOL_SRC_LABELS:
			return valCell(fd.SourceLabels, 1, 2)
		case DTLCOL_DST_LABELS:
			return valCell(fd.DestLabels, 1, 2)
		case DTLCOL_REPORTER:
			return valCell(fd.Reporter, 1, 0)
		case DTLCOL_PACK_IN:
			return valCell(intos(fd.PacketsIn), 1, 1)
		case DTLCOL_PACK_OUT:
			return valCell(intos(fd.PacketsOut), 1, 1)
		case DTLCOL_BYTE_IN:
			return valCell(intos(fd.BytesIn), 1, 1)
		case DTLCOL_BYTE_OUT:
			return valCell(intos(fd.BytesOut), 1, 1)
		case DTLCOL_ACTION:
			return valCell(fd.Action, 1, 0)
		}
	}
	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (fdt *flowSumDetailTable) GetRowCount() int {
	return fdt.fds.GetFlowDetailCount(fdt.key) + 1
}

func (fdt *flowSumDetailTable) GetColumnCount() int {
	return len(dtlCols)
}
