package tui

import (
	"fmt"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/rivo/tview"
)

// FlowDetailTableHeader is a table for displaying flow details header
type flowDetailHeaderTable struct {
	tview.TableContentReadOnly
	fd      *flowdata.FlowData
	allCols []string
}

func NewFlowDetailHeaderTable(fd *flowdata.FlowData) *flowDetailHeaderTable {
	return &flowDetailHeaderTable{
		fd:      fd,
		allCols: []string{"SRC NAMESPACE / NAME", "DST NAMESPACE / NAME", "RPT / PROTO:PORT", "START TIME", "END TIME", "P I/O - B I/O", "ACTION"},
	}
}

func (fdt *flowDetailHeaderTable) GetCell(row, column int) *tview.TableCell {
	const (
		SRC_NAMESPACE_NAME = iota
		DST_NAMESPACE_NAME
		RPT_PROTO_PORT
		START_TIME
		END_TIME
		PACK_IN_OUT_BYTE_IN_OUT
		ACTION
	)

	if row == 0 {
		return hdrCell(fdt.allCols[column], 1, 1)
	}

	fd := fdt.fd

	switch column {
	case SRC_NAMESPACE_NAME:
		tc := valCell(fmt.Sprintf("%s / %s", fd.SourceNamespace, fd.SourceName), 4, 3)
		tc.SetReference(fd.ID)
		return tc
	case DST_NAMESPACE_NAME:
		tc := valCell(fmt.Sprintf("%s / %s", fd.DestNamespace, fd.DestName), 4, 3)
		return tc
	case RPT_PROTO_PORT:
		return valCell(fmt.Sprintf("%s / %s:%d", fd.Reporter, fd.Protocol, fd.DestPort), 1, 0)
	case START_TIME:
		return valCell(tf(fd.StartTime), 1, 0)
	case END_TIME:
		return valCell(tf(fd.EndTime), 1, 0)
	case PACK_IN_OUT_BYTE_IN_OUT:
		return valCell(fmt.Sprintf("%d / %d - %d / %d", fd.PacketsIn, fd.PacketsOut, fd.BytesIn, fd.BytesOut), 1, 0)
	case ACTION:
		return actionCell(fd.Action)
	}
	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (fdt *flowDetailHeaderTable) GetRowCount() int {
	return 2
}

func (fdt *flowDetailHeaderTable) GetColumnCount() int {
	return len(fdt.allCols)
}
