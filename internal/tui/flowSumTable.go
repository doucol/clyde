package tui

import (
	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/rivo/tview"
)

type flowSumTable struct {
	tview.TableContentReadOnly
	fc        *flowcache.FlowCache
	fss       []*flowdata.FlowSum
	colTitles []string
}

const (
	tblcol_fst_SRC_NAMESPACE = iota
	tblcol_fst_SRC_NAME
	tblcol_fst_DST_NAMESPACE
	tblcol_fst_DST_NAME
	tblcol_fst_PROTO
	tblcol_fst_PORT
	tblcol_fst_SRC_COUNT
	tblcol_fst_DST_COUNT
	tblcol_fst_SRC_PACK_IN
	tblcol_fst_SRC_PACK_OUT
	tblcol_fst_SRC_BYTE_IN
	tblcol_fst_SRC_BYTE_OUT
	tblcol_fst_DST_PACK_IN
	tblcol_fst_DST_PACK_OUT
	tblcol_fst_DST_BYTE_IN
	tblcol_fst_DST_BYTE_OUT
	tblcol_fst_ACTION
)

func newFlowSumTable(fc *flowcache.FlowCache) *flowSumTable {
	return &flowSumTable{
		fc: fc,
		colTitles: []string{
			"SRC NAMESPACE",
			"SRC NAME",
			"DST NAMESPACE",
			"DST NAME",
			"PROTO",
			"PORT",
			"SRC COUNT",
			"DST COUNT",
			"SRC PACK IN",
			"SRC PACK OUT",
			"SRC BYTE IN",
			"SRC BYTE OUT",
			"DST PACK IN",
			"DST PACK OUT",
			"DST BYTE IN",
			"DST BYTE OUT",
			"ACTION",
		},
	}
}

func (fst *flowSumTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return hdrCell(fst.colTitles[column], 1, 1)
	}

	fs := fst.fss[row-1]

	switch column {
	case tblcol_fst_SRC_NAMESPACE:
		tc := valCell(fs.SourceNamespace, 1, 1)
		tc.SetReference(fs.ID)
		return tc
	case tblcol_fst_SRC_NAME:
		return valCell(fs.SourceName, 1, 2)
	case tblcol_fst_DST_NAMESPACE:
		return valCell(fs.DestNamespace, 1, 1)
	case tblcol_fst_DST_NAME:
		return valCell(fs.DestName, 1, 2)
	case tblcol_fst_PROTO:
		return valCell(fs.Protocol, 1, 0)
	case tblcol_fst_PORT:
		return valCell(intos(fs.DestPort), 1, 0)
	case tblcol_fst_SRC_COUNT:
		return valCell(intos(fs.SourceReports), 1, 1)
	case tblcol_fst_DST_COUNT:
		return valCell(intos(fs.DestReports), 1, 1)
	case tblcol_fst_SRC_PACK_IN:
		return valCell(uintos(fs.SourcePacketsIn), 1, 1)
	case tblcol_fst_SRC_PACK_OUT:
		return valCell(uintos(fs.SourcePacketsOut), 1, 1)
	case tblcol_fst_SRC_BYTE_IN:
		return valCell(uintos(fs.SourceBytesIn), 1, 1)
	case tblcol_fst_SRC_BYTE_OUT:
		return valCell(uintos(fs.SourceBytesOut), 1, 1)
	case tblcol_fst_DST_PACK_IN:
		return valCell(uintos(fs.DestPacketsIn), 1, 1)
	case tblcol_fst_DST_PACK_OUT:
		return valCell(uintos(fs.DestPacketsOut), 1, 1)
	case tblcol_fst_DST_BYTE_IN:
		return valCell(uintos(fs.DestBytesIn), 1, 1)
	case tblcol_fst_DST_BYTE_OUT:
		return valCell(uintos(fs.DestBytesOut), 1, 1)
	case tblcol_fst_ACTION:
		return actionCell(fs.Action)
	}

	return nil
}

func (fst *flowSumTable) GetRowCount() int {
	fst.fss = fst.fc.GetFlowSums()
	// Need to add 1 to account for the header row
	return len(fst.fss) + 1
}

func (fst *flowSumTable) GetColumnCount() int {
	return len(fst.colTitles)
}
