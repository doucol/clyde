package tui

import (
	"fmt"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/rivo/tview"
)

// FlowDetailTableHeader is a table for displaying flow details header
type flowDetailHeaderTable struct {
	tview.TableContentReadOnly
	fds       *flowdata.FlowDataStore
	fd        *flowdata.FlowData
	fas       *flowAppState
	colTitles []string
}

const (
	tblcol_fdht_SRC_NAMESPACE_NAME = iota
	tblcol_fdht_DST_NAMESPACE_NAME
	tblcol_fdht_RPT_PROTO_PORT
	tblcol_fdht_START_TIME
	tblcol_fdht_END_TIME
	tblcol_fdht_PACK_IN_OUT_BYTE_IN_OUT
	tblcol_fdht_ACTION
)

func newFlowDetailHeaderTable(fds *flowdata.FlowDataStore, fas *flowAppState) *flowDetailHeaderTable {
	return &flowDetailHeaderTable{
		fds: fds,
		fas: fas,
		colTitles: []string{
			"SRC NAMESPACE / NAME",
			"DST NAMESPACE / NAME",
			"RPT / PROTO:PORT",
			"START TIME",
			"END TIME",
			"P I/O - B I/O",
			"ACTION",
		},
	}
}

func (fdt *flowDetailHeaderTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return hdrCell(fdt.colTitles[column], 1, 1)
	}

	fd := fdt.fd

	switch column {
	case tblcol_fdht_SRC_NAMESPACE_NAME:
		tc := valCell(fmt.Sprintf("%s / %s", fd.SourceNamespace, fd.SourceName), 4, 3)
		tc.SetReference(fd.ID)
		return tc
	case tblcol_fdht_DST_NAMESPACE_NAME:
		tc := valCell(fmt.Sprintf("%s / %s", fd.DestNamespace, fd.DestName), 4, 3)
		return tc
	case tblcol_fdht_RPT_PROTO_PORT:
		return valCell(fmt.Sprintf("%s / %s:%d", fd.Reporter, fd.Protocol, fd.DestPort), 1, 0)
	case tblcol_fdht_START_TIME:
		return valCell(tf(fd.StartTime), 1, 0)
	case tblcol_fdht_END_TIME:
		return valCell(tf(fd.EndTime), 1, 0)
	case tblcol_fdht_PACK_IN_OUT_BYTE_IN_OUT:
		return valCell(fmt.Sprintf("%d / %d - %d / %d", fd.PacketsIn, fd.PacketsOut, fd.BytesIn, fd.BytesOut), 1, 0)
	case tblcol_fdht_ACTION:
		return actionCell(fd.Action)
	}
	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (fdt *flowDetailHeaderTable) GetRowCount() int {
	if fdt.fas.flowID <= 0 {
		return 1
	}
	fdt.fd = fdt.fds.GetFlowDetail(fdt.fas.flowID)
	return 2
}

func (fdt *flowDetailHeaderTable) GetColumnCount() int {
	return len(fdt.colTitles)
}
