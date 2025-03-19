package tui

import (
	"fmt"
	"strings"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/rivo/tview"
)

// FlowDetailTableHeader is a table for displaying flow details header
type flowKeyHeaderTable struct {
	tview.TableContentReadOnly
	fds *flowdata.FlowDataStore
	key string
}

func (fdt *flowKeyHeaderTable) GetCell(row, column int) *tview.TableCell {
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

func (fdt *flowKeyHeaderTable) GetRowCount() int {
	return 2
}

func (fdt *flowKeyHeaderTable) GetColumnCount() int {
	return len(keyCols)
}
