package whisker

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	keyCols    = []string{"SRC NAMESPACE", "SRC NAME", "DST NAMESPACE", "DST NAME", "PROTO", "PORT"}
	datCols    = []string{"SRC COUNT", "DST COUNT", "SRC PACK IN", "SRC PACK OUT", "SRC BYTE IN", "SRC BYTE OUT", "DST PACK IN", "DST PACK OUT", "DST BYTE IN", "DST BYTE OUT", "ACTION"}
	sumCols    = append(keyCols, datCols...)
	dtlCols    = []string{"START TIME", "END TIME", "SRC LABELS", "DST LABELS", "REPORTER", "PACK IN", "PACK OUT", "BYTE IN", "BYTE OUT", "ACTION"}
	titleStyle = tcell.Style{}.Bold(true)
)

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
		return tview.NewTableCell(sumCols[column]).SetMaxWidth(1).SetExpansion(1).SetStyle(titleStyle)
	}
	if fs, ok := fst.fds.GetFlowSum(row); ok {
		switch column {
		case 0:
			return tview.NewTableCell(fs.SourceNamespace).SetMaxWidth(1).SetExpansion(1)
		case 1:
			return tview.NewTableCell(fs.SourceName).SetMaxWidth(1).SetExpansion(2)
		case 2:
			return tview.NewTableCell(fs.DestNamespace).SetMaxWidth(1).SetExpansion(1)
		case 3:
			return tview.NewTableCell(fs.DestName).SetMaxWidth(1).SetExpansion(2)
		case 4:
			return tview.NewTableCell(fs.Protocol).SetMaxWidth(1).SetExpansion(0)
		case 5:
			return tview.NewTableCell(intos(fs.DestPort)).SetMaxWidth(1).SetExpansion(0)
		case 6:
			return tview.NewTableCell(intos(fs.SourceReports)).SetMaxWidth(1).SetExpansion(1)
		case 7:
			return tview.NewTableCell(intos(fs.DestReports)).SetMaxWidth(1).SetExpansion(1)
		case 8:
			return tview.NewTableCell(uintos(fs.SourcePacketsIn)).SetMaxWidth(1).SetExpansion(1)
		case 9:
			return tview.NewTableCell(uintos(fs.SourcePacketsOut)).SetMaxWidth(1).SetExpansion(1)
		case 10:
			return tview.NewTableCell(uintos(fs.SourceBytesIn)).SetMaxWidth(1).SetExpansion(1)
		case 11:
			return tview.NewTableCell(uintos(fs.SourceBytesOut)).SetMaxWidth(1).SetExpansion(1)
		case 12:
			return tview.NewTableCell(uintos(fs.DestPacketsIn)).SetMaxWidth(1).SetExpansion(1)
		case 13:
			return tview.NewTableCell(uintos(fs.DestPacketsOut)).SetMaxWidth(1).SetExpansion(1)
		case 14:
			return tview.NewTableCell(uintos(fs.DestBytesIn)).SetMaxWidth(1).SetExpansion(1)
		case 15:
			return tview.NewTableCell(uintos(fs.DestBytesOut)).SetMaxWidth(1).SetExpansion(1)
		case 16:
			return tview.NewTableCell(fs.Action).SetMaxWidth(1).SetExpansion(0)
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
		return tview.NewTableCell(dtlCols[column]).SetMaxWidth(1).SetExpansion(1).SetStyle(titleStyle)
	}
	if fd, ok := fdt.fds.GetFlowDetail(fdt.key, row); ok {
		switch column {
		case 0:
			return tview.NewTableCell(fd.StartTime.Format(time.RFC3339)).SetMaxWidth(1).SetExpansion(1)
		case 1:
			return tview.NewTableCell(fd.EndTime.Format(time.RFC3339)).SetMaxWidth(1).SetExpansion(2)
		case 2:
			return tview.NewTableCell(fd.SourceLabels).SetMaxWidth(1).SetExpansion(1)
		case 3:
			return tview.NewTableCell(fd.DestLabels).SetMaxWidth(1).SetExpansion(1)
		case 4:
			return tview.NewTableCell(fd.Reporter).SetMaxWidth(1).SetExpansion(0)
		case 5:
			return tview.NewTableCell(intos(fd.PacketsIn)).SetMaxWidth(1).SetExpansion(1)
		case 6:
			return tview.NewTableCell(intos(fd.PacketsOut)).SetMaxWidth(1).SetExpansion(1)
		case 7:
			return tview.NewTableCell(intos(fd.BytesIn)).SetMaxWidth(1).SetExpansion(1)
		case 8:
			return tview.NewTableCell(intos(fd.BytesOut)).SetMaxWidth(1).SetExpansion(1)
		case 9:
			return tview.NewTableCell(fd.Action).SetMaxWidth(1).SetExpansion(0)
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
		return tview.NewTableCell(keyCols[column]).SetMaxWidth(1).SetExpansion(1).SetStyle(titleStyle)
	}
	keyVals := strings.Split(fdt.key, "|")
	switch column {
	case 0:
		return tview.NewTableCell(keyVals[column]).SetMaxWidth(1).SetExpansion(1)
	case 1:
		return tview.NewTableCell(keyVals[column]).SetMaxWidth(1).SetExpansion(2)
	case 2:
		return tview.NewTableCell(keyVals[column]).SetMaxWidth(1).SetExpansion(1)
	case 3:
		return tview.NewTableCell(keyVals[column]).SetMaxWidth(1).SetExpansion(2)
	case 4:
		return tview.NewTableCell(keyVals[column]).SetMaxWidth(1).SetExpansion(0)
	case 5:
		return tview.NewTableCell(keyVals[column]).SetMaxWidth(1).SetExpansion(0)
	}
	panic(fmt.Errorf("invalid cell row: %d, col: %d", row, column))
}

func (fdt *flowDetailTableHeader) GetRowCount() int {
	return 2
}

func (fdt *flowDetailTableHeader) GetColumnCount() int {
	return len(keyCols)
}
