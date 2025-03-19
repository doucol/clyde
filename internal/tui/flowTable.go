package tui

import (
	"strconv"

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

func uintos(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func intos(v int64) string {
	return strconv.FormatInt(v, 10)
}
