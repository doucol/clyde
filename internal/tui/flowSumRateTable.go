package tui

import (
	"fmt"

	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/rivo/tview"
)

type flowSumRateTable struct {
	tview.TableContentReadOnly
	fc        *flowcache.FlowCache
	fss       []*flowdata.FlowSum
	colTitles []string
}

const (
	tblcol_fsrt_SRC_NAMESPACE_NAME = iota
	tblcol_fsrt_DST_NAMESPACE_NAME
	tblcol_fsrt_PROTO_PORT
	tblcol_fsrt_SRC_PACK_RATE
	tblcol_fsrt_SRC_BYTE_RATE
	tblcol_fsrt_DST_PACK_RATE
	tblcol_fsrt_DST_BYTE_RATE
	tblcol_fsrt_ACTION
)

func newFlowSumRateTable(fc *flowcache.FlowCache) *flowSumRateTable {
	return &flowSumRateTable{
		fc: fc,
		colTitles: []string{
			"SRC NAMESPACE / NAME",
			"DST NAMESPACE / NAME",
			"PROTO:PORT",
			"SRC PACK/SEC",
			"SRC BYTE/SEC",
			"DST PACK/SEC",
			"DST BYTE/SEC",
			"ACTION",
		},
	}
}

func (t *flowSumRateTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		tc := hdrCell(t.colTitles[column], 1, 1)
		tc.SetReference(0)
		return tc
	}

	fs := t.fss[row-1]

	switch column {
	case tblcol_fsrt_SRC_NAMESPACE_NAME:
		tc := valCell(fmt.Sprintf("%s / %s", fs.SourceNamespace, fs.SourceName), 4, 3)
		tc.SetReference(fs.ID)
		return tc
	case tblcol_fsrt_DST_NAMESPACE_NAME:
		tc := valCell(fmt.Sprintf("%s / %s", fs.DestNamespace, fs.DestName), 4, 3)
		return tc
	case tblcol_fsrt_PROTO_PORT:
		return valCell(fmt.Sprintf("%s:%d", fs.Protocol, fs.DestPort), 1, 0)
	case tblcol_fsrt_SRC_PACK_RATE:
		return valCell(fmt.Sprintf("%.2f", fs.SourceTotalPacketRate), 1, 1)
	case tblcol_fsrt_SRC_BYTE_RATE:
		return valCell(fmt.Sprintf("%.2f", fs.SourceTotalByteRate), 1, 1)
	case tblcol_fsrt_DST_PACK_RATE:
		return valCell(fmt.Sprintf("%.2f", fs.DestTotalPacketRate), 1, 1)
	case tblcol_fsrt_DST_BYTE_RATE:
		return valCell(fmt.Sprintf("%.2f", fs.DestTotalByteRate), 1, 1)
	case tblcol_fsrt_ACTION:
		return actionCell(fs.Action)
	}

	return nil
}

func (t *flowSumRateTable) GetRowCount() int {
	t.fss = t.fc.GetFlowSumRates()
	return len(t.fss) + 1
}

func (t *flowSumRateTable) GetColumnCount() int {
	return len(t.colTitles)
}
