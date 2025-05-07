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
	fc        *flowcache.FlowCache
	flows     []*flowdata.FlowData
	fas       *flowAppState
	colTitles []string
}

const (
	tblcol_fsdt_START_TIME = iota
	tblcol_fsdt_END_TIME
	tblcol_fsdt_SRC_LABELS
	tblcol_fsdt_DST_LABELS
	tblcol_fsdt_REPORTER
	tblcol_fsdt_PACK_IN
	tblcol_fsdt_PACK_OUT
	tblcol_fsdt_BYTE_IN
	tblcol_fsdt_BYTE_OUT
	tblcol_fsdt_ACTION
)

func newFlowSumDetailTable(fc *flowcache.FlowCache, fas *flowAppState) *flowSumDetailTable {
	return &flowSumDetailTable{
		fc:  fc,
		fas: fas,
		colTitles: []string{
			"START TIME",
			"END TIME",
			"SRC LABELS",
			"DST LABELS",
			"REPORTER",
			"PACK IN",
			"PACK OUT",
			"BYTE IN",
			"BYTE OUT",
			"ACTION",
		},
	}
}

func (fdt *flowSumDetailTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		tc := hdrCell(fdt.colTitles[column], 1, 1)
		tc.SetReference(0)
		return tc
	}

	fd := fdt.flows[row-1]

	switch column {
	case tblcol_fsdt_START_TIME:
		tc := valCell(tf(fd.StartTime), 3, 0)
		tc.SetReference(fd.ID)
		return tc
	case tblcol_fsdt_END_TIME:
		return valCell(tf(fd.EndTime), 3, 0)
	case tblcol_fsdt_SRC_LABELS:
		return valCell(fd.SourceLabels, 2, 3)
	case tblcol_fsdt_DST_LABELS:
		return valCell(fd.DestLabels, 2, 3)
	case tblcol_fsdt_REPORTER:
		return valCell(fd.Reporter, 1, 0)
	case tblcol_fsdt_PACK_IN:
		return valCell(intos(fd.PacketsIn), 1, 0)
	case tblcol_fsdt_PACK_OUT:
		return valCell(intos(fd.PacketsOut), 1, 0)
	case tblcol_fsdt_BYTE_IN:
		return valCell(intos(fd.BytesIn), 1, 0)
	case tblcol_fsdt_BYTE_OUT:
		return valCell(intos(fd.BytesOut), 1, 0)
	case tblcol_fsdt_ACTION:
		return actionCell(fd.Action)
	}

	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (fdt *flowSumDetailTable) GetRowCount() int {
	id := 0
	switch fdt.fas.lastHomePage {
	case pageSummaryTotalsName:
		id = fdt.fas.sumID
	case pageSummaryRatesName:
		id = fdt.fas.rateID
	default:
		panic(fmt.Errorf("flowKeyHeaderTable: invalid lastHomePage: %s", fdt.fas.lastHomePage))
	}
	if id <= 0 {
		return 1
	}
	fdt.flows = fdt.fc.GetFlowsBySumID(id)
	return len(fdt.flows) + 1
}

func (fdt *flowSumDetailTable) GetColumnCount() int {
	return len(fdt.colTitles)
}
