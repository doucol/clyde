package tui

import (
	"fmt"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/rivo/tview"
)

// FlowKeyHeaderTable is a table for displaying flow details header
type flowKeyHeaderTable struct {
	tview.TableContentReadOnly
	fds       *flowdata.FlowDataStore
	fs        *flowdata.FlowSum
	fas       *flowAppState
	colTitles []string
}

const (
	tblcol_fkht_SRC_NAMESPACE = iota
	tblcol_fkht_SRC_NAME
	tblcol_fkht_DST_NAMESPACE
	tblcol_fkht_DST_NAME
	tblcol_fkht_PROTO
	tblcol_fkht_PORT
)

func newFlowKeyHeaderTable(fds *flowdata.FlowDataStore, fas *flowAppState) *flowKeyHeaderTable {
	return &flowKeyHeaderTable{
		fds: fds,
		fas: fas,
		colTitles: []string{
			"SRC NAMESPACE",
			"SRC NAME",
			"DST NAMESPACE",
			"DST NAME",
			"PROTO",
			"PORT",
		},
	}
}

func (fdt *flowKeyHeaderTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return hdrCell(fdt.colTitles[column], 1, 1)
	}
	switch column {
	case tblcol_fkht_SRC_NAMESPACE:
		return valCell(fdt.fs.SourceNamespace, 1, 1)
	case tblcol_fkht_SRC_NAME:
		return valCell(fdt.fs.SourceName, 1, 2)
	case tblcol_fkht_DST_NAMESPACE:
		return valCell(fdt.fs.DestNamespace, 1, 1)
	case tblcol_fkht_DST_NAME:
		return valCell(fdt.fs.DestName, 1, 2)
	case tblcol_fkht_PROTO:
		return valCell(fdt.fs.Protocol, 1, 0)
	case tblcol_fkht_PORT:
		return valCell(intos(fdt.fs.DestPort), 1, 0)
	}
	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (fdt *flowKeyHeaderTable) GetRowCount() int {
	fdt.fs = fdt.fds.GetFlowSum(fdt.fas.SumID)
	if fdt.fs == nil {
		panic(fmt.Errorf("flowKeyHeaderTable: flowSum with ID %d not found", fdt.fas.SumID))
	}
	return 2
}

func (fdt *flowKeyHeaderTable) GetColumnCount() int {
	return len(fdt.colTitles)
}
