package tui

import (
	"strings"

	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type flowSumTable struct {
	tview.TableContentReadOnly
	fc  *flowcache.FlowCache
	fss []*flowdata.FlowSum
}

func (fst *flowSumTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return hdrCell(sumCols[column], 1, 1)
	}

	fs := fst.fss[row-1]

	switch column {
	case SUMCOL_SRC_NAMESPACE:
		tc := valCell(fs.SourceNamespace, 1, 1)
		tc.SetReference(fs.ID)
		return tc
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
		tc := valCell(fs.Action, 1, 0)
		setSumCellStyle(fs, tc)
		return tc
	}

	return nil
}

func setSumCellStyle(fs *flowdata.FlowSum, tc *tview.TableCell) {
	color := tcell.ColorLightSkyBlue
	if strings.ToLower(fs.Action) == "deny" {
		color = tcell.ColorOrangeRed
	}
	tc.SetTextColor(color)
	tc.SetSelectedStyle(selectedStyle.Foreground(color))
}

func (fst *flowSumTable) GetRowCount() int {
	fst.fss = fst.fc.GetFlowSums()
	// Need to add 1 to account for the header row
	return len(fst.fss) + 1
}

func (fst *flowSumTable) GetColumnCount() int {
	return len(sumCols)
}
