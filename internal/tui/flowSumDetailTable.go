package tui

import (
	"fmt"

	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/rivo/tview"
)

// FlowDetailTable is a table for displaying flow details.
type flowSumDetailTable struct {
	tview.TableContentReadOnly
	fc    *flowcache.FlowCache
	flows []*flowdata.FlowData
	sumID int
}

func (fdt *flowSumDetailTable) CurrentID() {
}

func (fdt *flowSumDetailTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return hdrCell(dtlCols[column], 1, 1)
	}

	fd := fdt.flows[row-1]

	switch column {
	case DTLCOL_START_TIME:
		tc := valCell(tf(fd.StartTime), 3, 0)
		tc.SetReference(fd.ID)
		return tc
	case DTLCOL_END_TIME:
		return valCell(tf(fd.EndTime), 3, 0)
	case DTLCOL_SRC_LABELS:
		return valCell(fd.SourceLabels, 2, 3)
	case DTLCOL_DST_LABELS:
		return valCell(fd.DestLabels, 2, 3)
	case DTLCOL_REPORTER:
		return valCell(fd.Reporter, 1, 0)
	case DTLCOL_PACK_IN:
		return valCell(intos(fd.PacketsIn), 1, 0)
	case DTLCOL_PACK_OUT:
		return valCell(intos(fd.PacketsOut), 1, 0)
	case DTLCOL_BYTE_IN:
		return valCell(intos(fd.BytesIn), 1, 0)
	case DTLCOL_BYTE_OUT:
		return valCell(intos(fd.BytesOut), 1, 0)
	case DTLCOL_ACTION:
		return actionCell(fd.Action)
	}

	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (fdt *flowSumDetailTable) GetRowCount() int {
	fdt.flows = fdt.fc.GetFlowsBySumID(fdt.sumID)
	return len(fdt.flows) + 1
}

func (fdt *flowSumDetailTable) GetColumnCount() int {
	return len(dtlCols)
}
