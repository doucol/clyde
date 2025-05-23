package tui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var localLoc *time.Location

func init() {
	var err error
	localLoc, err = time.LoadLocation("Local")
	if err != nil {
		panic(errors.New("unable to load 'local' time location: " + err.Error()))
	}
}

var (
	bgColor            = tcell.ColorBlack
	textColor          = tcell.ColorLightGray
	borderColor        = tcell.ColorDarkSlateGray
	titleColor         = tcell.ColorWhite
	allowColor         = tcell.ColorLightSkyBlue
	denyColor          = tcell.ColorOrangeRed
	selBgColor         = tcell.ColorDarkBlue
	formBgColor        = tcell.ColorLightSlateGrey
	formFgColor        = tcell.ColorBlack
	selFgColor         = titleColor
	selectedStyle      = tcell.StyleDefault.Foreground(selFgColor).Background(selBgColor)
	tcellValStyle      = tcell.StyleDefault.Background(bgColor).Foreground(textColor)
	tcellHdrStyle      = tcellValStyle.Bold(true).Underline(false)
	formLabelStyle     = tcell.StyleDefault.Bold(true).Underline(false)
	listUnselStyle     = tcell.StyleDefault.Foreground(selFgColor).Background(selBgColor)
	listSelStyle       = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(selBgColor)
	formSelBorderStyle = selectedStyle
)

func setTheme() {
	tview.Styles.PrimitiveBackgroundColor = bgColor
	tview.Styles.ContrastBackgroundColor = bgColor
	tview.Styles.MoreContrastBackgroundColor = bgColor
	tview.Styles.PrimaryTextColor = titleColor
	tview.Styles.BorderColor = borderColor
	tview.Styles.TitleColor = titleColor
	tview.Styles.GraphicsColor = titleColor
	tview.Styles.SecondaryTextColor = titleColor
	tview.Styles.TertiaryTextColor = titleColor
	tview.Styles.InverseTextColor = titleColor
	tview.Styles.ContrastSecondaryTextColor = titleColor
}

func applyTheme(components ...tview.Primitive) {
	for _, p := range components {
		switch c := p.(type) {
		case *tview.Flex:
			c.SetBackgroundColor(bgColor)
			c.SetTitleColor(titleColor)
			c.SetBorderColor(borderColor)
		case *tview.Table:
			c.SetBackgroundColor(bgColor)
			c.SetBorderColor(borderColor)
			c.SetTitleColor(titleColor)
			c.SetSelectedStyle(selectedStyle)
		case *tview.TextView:
			c.SetBackgroundColor(bgColor)
			c.SetTextColor(textColor)
			c.SetBorderColor(borderColor)
			c.SetTitleColor(titleColor)
		case *tview.Box:
			c.SetBackgroundColor(bgColor)
			c.SetBorderColor(borderColor)
			c.SetTitleColor(titleColor)
		case *tview.DropDown:
			c.SetBackgroundColor(bgColor)
			c.SetBorderColor(borderColor)
			c.SetTitleColor(titleColor)
			c.SetFieldBackgroundColor(selBgColor)
			c.SetFieldTextColor(selFgColor)
			c.SetLabelColor(titleColor)
			c.SetLabelStyle(formLabelStyle)
			c.SetListStyles(listUnselStyle, listSelStyle)
		case *tview.InputField:
			c.SetBackgroundColor(selBgColor)
			c.SetBorderColor(borderColor)
			c.SetTitleColor(titleColor)
			c.SetFieldBackgroundColor(selBgColor)
			c.SetFieldTextColor(selFgColor)
			c.SetLabelStyle(formLabelStyle)
			c.SetLabelColor(titleColor)
		case *tview.Form:
			c.SetBackgroundColor(formBgColor)
			c.SetBorderColor(borderColor)
			c.SetTitleColor(formFgColor)
			c.SetFieldBackgroundColor(selBgColor)
			c.SetFieldTextColor(selFgColor)
			c.SetLabelColor(formFgColor)
			c.SetFieldStyle(selectedStyle)
		}
	}
}

func newTable() *tview.Table {
	t := tview.NewTable().SetBorders(false).SetSelectable(false, false)
	applyTheme(t)
	return t
}

func cell(val string, width, exp int) *tview.TableCell {
	return tview.NewTableCell(val).SetMaxWidth(width).SetExpansion(exp)
}

func valCell(val string, width, exp int) *tview.TableCell {
	return cell(val, width, exp).SetStyle(tcellValStyle)
}

func hdrCell(val string, width, exp int) *tview.TableCell {
	return cell(val, width, exp).SetStyle(tcellHdrStyle)
}

func actionCell(action string) *tview.TableCell {
	tc := valCell(action, 1, 0)
	color := allowColor
	if strings.ToLower(action) == "deny" {
		color = denyColor
	}
	tc.SetTextColor(color)
	tc.SetSelectedStyle(selectedStyle.Foreground(color))
	return tc
}

//	func uintos(v uint64) string {
//		return strconv.FormatUint(v, 10)
//	}
func intos(v int64) string {
	return strconv.FormatInt(v, 10)
}

func tf(t time.Time) string {
	// TODO: This turned out to be too long, need to format diff
	// return t.In(localLoc).Format(time.RFC3339)
	return t.Format(time.RFC3339)
}

func policyHitsToString(policyHits []*flowdata.PolicyHit) string {
	var s string
	for _, ph := range policyHits {
		s += fmt.Sprintf("%s\n", policyHitToString(ph))
	}
	return s
}

func policyHitToString(ph *flowdata.PolicyHit) string {
	phs := fmt.Sprintf("Kind: %s\nName: %s\nNamespace: %s\nTier: %s\nAction: %s\nPolicyIndex: %d\nRuleIndex: %d\n",
		ph.Kind, ph.Name, ph.Namespace, ph.Tier, ph.Action, ph.PolicyIndex, ph.RuleIndex)
	if ph.Trigger != nil {
		phs += "\nTriggers:\n" + policyHitTriggerToString(ph.Trigger)
	}
	return phs
}

func policyHitTriggerToString(ph *flowdata.PolicyHit) string {
	phs := fmt.Sprintf("\tKind: %s\n\tName: %s\n\tNamespace: %s\n\tTier: %s\n\tAction: %s\n\tPolicyIndex: %d\n\tRuleIndex: %d\n",
		ph.Kind, ph.Name, ph.Namespace, ph.Tier, ph.Action, ph.PolicyIndex, ph.RuleIndex)
	if ph.Trigger != nil {
		phs += "\n\n" + policyHitTriggerToString(ph.Trigger)
	}
	return phs
}
