package tui

import (
	"fmt"

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
	tblcol_fst_SRC_NAMESPACE_NAME = iota
	tblcol_fst_DST_NAMESPACE_NAME
	tblcol_fst_PROTO_PORT
	tblcol_fst_SRC_DST_COUNT
	tblcol_fst_SRC_PACK_IO
	tblcol_fst_SRC_BYTE_IO
	tblcol_fst_DST_PACK_IO
	tblcol_fst_DST_BYTE_IO
	tblcol_fst_ACTION
)

func newFlowSumTable(fc *flowcache.FlowCache) *flowSumTable {
	return &flowSumTable{
		fc: fc,
		colTitles: []string{
			"SRC NAMESPACE / NAME",
			"DST NAMESPACE / NAME",
			"PROTO:PORT",
			"SRC / DST",
			"SRC PACK I/O",
			"SRC BYTE I/O",
			"DST PACK I/O",
			"DST BYTE I/O",
			"ACTION",
		},
	}
}

func (fst *flowSumTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		tc := hdrCell(fst.colTitles[column], 1, 1)
		tc.SetReference(0)
		return tc
	}

	fs := fst.fss[row-1]

	switch column {
	case tblcol_fst_SRC_NAMESPACE_NAME:
		tc := valCell(fmt.Sprintf("%s / %s", fs.SourceNamespace, fs.SourceName), 4, 3)
		tc.SetReference(fs.ID)
		return tc
	case tblcol_fst_DST_NAMESPACE_NAME:
		tc := valCell(fmt.Sprintf("%s / %s", fs.DestNamespace, fs.DestName), 4, 3)
		return tc
	case tblcol_fst_PROTO_PORT:
		return valCell(fmt.Sprintf("%s:%d", fs.Protocol, fs.DestPort), 1, 0)
	case tblcol_fst_SRC_DST_COUNT:
		return valCell(fmt.Sprintf("%d / %d", fs.SourceReports, fs.DestReports), 1, 1)
	case tblcol_fst_SRC_PACK_IO:
		return valCell(fmt.Sprintf("%d / %d", fs.SourcePacketsIn, fs.SourcePacketsOut), 1, 2)
	case tblcol_fst_SRC_BYTE_IO:
		return valCell(fmt.Sprintf("%d / %d", fs.SourceBytesIn, fs.SourceBytesOut), 1, 2)
	case tblcol_fst_DST_PACK_IO:
		return valCell(fmt.Sprintf("%d / %d", fs.DestPacketsIn, fs.DestPacketsOut), 1, 2)
	case tblcol_fst_DST_BYTE_IO:
		return valCell(fmt.Sprintf("%d / %d", fs.DestBytesIn, fs.DestBytesOut), 1, 2)
	case tblcol_fst_ACTION:
		return actionCell(fs.Action)
	}

	return nil
}

func (fst *flowSumTable) GetRowCount() int {
	fst.fss = fst.fc.GetFlowSumTotals()
	return len(fst.fss) + 1
}

func (fst *flowSumTable) GetColumnCount() int {
	return len(fst.colTitles)
}
