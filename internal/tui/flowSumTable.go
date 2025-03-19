package tui

import (
	"fmt"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/rivo/tview"
)

type flowSumTable struct {
	tview.TableContentReadOnly
	fds *flowdata.FlowDataStore
}

func (fst *flowSumTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return hdrCell(sumCols[column], 1, 1)
	}
	if fs, ok := fst.fds.GetFlowSum(row); ok {
		switch column {
		case SUMCOL_SRC_NAMESPACE:
			return valCell(fs.SourceNamespace, 1, 1)
		case SUMCOL_SRC_NAME:
			return valCell(fs.SourceName, 1, 2)
		case SUMCOL_DST_NAMESPACE:
			return valCell(fs.DestNamespace, 1, 1)
		case SUMCOL_DST_NAME:
			return valCell(fs.DestName, 1, 2)
		case SUMCOL_PROTO:
			return valCell(fs.Protocol, 1, 0)
		case SUMCOL_PORT:
			return valCell(intos(fs.DestPort), 1, 0)
		case SUMCOL_SRC_COUNT:
			return valCell(intos(fs.SourceReports), 1, 1)
		case SUMCOL_DST_COUNT:
			return valCell(intos(fs.DestReports), 1, 1)
		case SUMCOL_SRC_PACK_IN:
			return valCell(uintos(fs.SourcePacketsIn), 1, 1)
		case SUMCOL_SRC_PACK_OUT:
			return valCell(uintos(fs.SourcePacketsOut), 1, 1)
		case SUMCOL_SRC_BYTE_IN:
			return valCell(uintos(fs.SourceBytesIn), 1, 1)
		case SUMCOL_SRC_BYTE_OUT:
			return valCell(uintos(fs.SourceBytesOut), 1, 1)
		case SUMCOL_DST_PACK_IN:
			return valCell(uintos(fs.DestPacketsIn), 1, 1)
		case SUMCOL_DST_PACK_OUT:
			return valCell(uintos(fs.DestPacketsOut), 1, 1)
		case SUMCOL_DST_BYTE_IN:
			return valCell(uintos(fs.DestBytesIn), 1, 1)
		case SUMCOL_DST_BYTE_OUT:
			return valCell(uintos(fs.DestBytesOut), 1, 1)
		case SUMCOL_ACTION:
			return valCell(fs.Action, 1, 0)
		}
	}
	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (fst *flowSumTable) GetRowCount() int {
	return fst.fds.GetFlowSumCount() + 1
}

func (fst *flowSumTable) GetColumnCount() int {
	return len(sumCols)
}
