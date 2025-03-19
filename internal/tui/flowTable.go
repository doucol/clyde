package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	SUMCOL_SRC_NAMESPACE = iota
	SUMCOL_SRC_NAME
	SUMCOL_DST_NAMESPACE
	SUMCOL_DST_NAME
	SUMCOL_PROTO
	SUMCOL_PORT
	SUMCOL_SRC_COUNT
	SUMCOL_DST_COUNT
	SUMCOL_SRC_PACK_IN
	SUMCOL_SRC_PACK_OUT
	SUMCOL_SRC_BYTE_IN
	SUMCOL_SRC_BYTE_OUT
	SUMCOL_DST_PACK_IN
	SUMCOL_DST_PACK_OUT
	SUMCOL_DST_BYTE_IN
	SUMCOL_DST_BYTE_OUT
	SUMCOL_ACTION
)

const (
	DTLCOL_START_TIME = iota
	DTLCOL_END_TIME
	DTLCOL_SRC_LABELS
	DTLCOL_DST_LABELS
	DTLCOL_REPORTER
	DTLCOL_PACK_IN
	DTLCOL_PACK_OUT
	DTLCOL_BYTE_IN
	DTLCOL_BYTE_OUT
	DTLCOL_ACTION
)

var (
	keyCols = []string{"SRC NAMESPACE", "SRC NAME", "DST NAMESPACE", "DST NAME", "PROTO", "PORT"}
	datCols = []string{"SRC COUNT", "DST COUNT", "SRC PACK IN", "SRC PACK OUT", "SRC BYTE IN", "SRC BYTE OUT", "DST PACK IN", "DST PACK OUT", "DST BYTE IN", "DST BYTE OUT", "ACTION"}
	sumCols = append(keyCols, datCols...)
	dtlCols = []string{"START TIME", "END TIME", "SRC LABELS", "DST LABELS", "REPORTER", "PACK IN", "PACK OUT", "BYTE IN", "BYTE OUT", "ACTION"}
	// hdrStyle = tcell.Style{}.Normal().Background(tcell.ColorBlack).Foreground(tcell.ColorWhite).Bold(true)
	// valStyle = tcell.Style{}.Normal().Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	hdrStyle = tcell.Style{}.Bold(true)
	valStyle = tcell.Style{}
)

func cell(val string, width, exp int) *tview.TableCell {
	return tview.NewTableCell(val).SetMaxWidth(width).SetExpansion(exp)
}

func valCell(val string, width, exp int) *tview.TableCell {
	return cell(val, width, exp).SetStyle(valStyle)
}

func hdrCell(val string, width, exp int) *tview.TableCell {
	return cell(val, width, exp).SetStyle(hdrStyle)
}

type flowSumTable struct {
	tview.TableContentReadOnly
	fds *flowdata.FlowDataStore
}

func uintos(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func intos(v int64) string {
	return strconv.FormatInt(v, 10)
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

// FlowDetailTable is a table for displaying flow details.
type flowDetailTable struct {
	tview.TableContentReadOnly
	fds *flowdata.FlowDataStore
	key string
}

func (fdt *flowDetailTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return hdrCell(dtlCols[column], 1, 1)
	}
	if fd, ok := fdt.fds.GetFlowDetail(fdt.key, row); ok {
		switch column {
		case DTLCOL_START_TIME:
			return valCell(fd.StartTime.Format(time.RFC3339), 1, 0)
		case DTLCOL_END_TIME:
			return valCell(fd.EndTime.Format(time.RFC3339), 1, 0)
		case DTLCOL_SRC_LABELS:
			return valCell(fd.SourceLabels, 1, 2)
		case DTLCOL_DST_LABELS:
			return valCell(fd.DestLabels, 1, 2)
		case DTLCOL_REPORTER:
			return valCell(fd.Reporter, 1, 0)
		case DTLCOL_PACK_IN:
			return valCell(intos(fd.PacketsIn), 1, 1)
		case DTLCOL_PACK_OUT:
			return valCell(intos(fd.PacketsOut), 1, 1)
		case DTLCOL_BYTE_IN:
			return valCell(intos(fd.BytesIn), 1, 1)
		case DTLCOL_BYTE_OUT:
			return valCell(intos(fd.BytesOut), 1, 1)
		case DTLCOL_ACTION:
			return valCell(fd.Action, 1, 0)
		}
	}
	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (fdt *flowDetailTable) GetRowCount() int {
	return fdt.fds.GetFlowDetailCount(fdt.key) + 1
}

func (fdt *flowDetailTable) GetColumnCount() int {
	return len(dtlCols)
}

// FlowDetailTableHeader is a table for displaying flow details header
type flowDetailTableHeader struct {
	tview.TableContentReadOnly
	fds *flowdata.FlowDataStore
	key string
}

func (fdt *flowDetailTableHeader) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return hdrCell(keyCols[column], 1, 1)
	}
	keyVals := strings.Split(fdt.key, "|")
	switch column {
	case SUMCOL_SRC_NAMESPACE:
		return valCell(keyVals[SUMCOL_SRC_NAMESPACE], 1, 1)
	case SUMCOL_SRC_NAME:
		return valCell(keyVals[SUMCOL_SRC_NAME], 1, 2)
	case SUMCOL_DST_NAMESPACE:
		return valCell(keyVals[SUMCOL_DST_NAMESPACE], 1, 1)
	case SUMCOL_DST_NAME:
		return valCell(keyVals[SUMCOL_DST_NAME], 1, 2)
	case SUMCOL_PROTO:
		return valCell(keyVals[SUMCOL_PROTO], 1, 0)
	case SUMCOL_PORT:
		return valCell(keyVals[SUMCOL_PORT], 1, 0)
	}
	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (fdt *flowDetailTableHeader) GetRowCount() int {
	return 2
}

func (fdt *flowDetailTableHeader) GetColumnCount() int {
	return len(keyCols)
}
