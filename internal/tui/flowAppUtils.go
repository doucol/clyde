package tui

import (
	"fmt"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	bgColor       = tcell.ColorBlack
	textColor     = tcell.ColorLightGray
	borderColor   = tcell.ColorDarkSlateGray
	titleColor    = tcell.ColorWhite
	allowColor    = tcell.ColorLightSkyBlue
	denyColor     = tcell.ColorOrangeRed
	selectedStyle = tcell.StyleDefault.Foreground(titleColor).Background(tcell.ColorDarkBlue)
	tcellValStyle = tcell.StyleDefault.Background(bgColor).Foreground(textColor)
	tcellHdrStyle = tcellValStyle.Bold(true).Underline(false)
)

func setTheme() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.BorderColor = tcell.ColorBlue
	tview.Styles.TitleColor = tcell.ColorWhite
	tview.Styles.GraphicsColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorWhite
	tview.Styles.TertiaryTextColor = tcell.ColorWhite
	tview.Styles.InverseTextColor = tcell.ColorWhite
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorWhite
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
		}
	}
}

func newTable() *tview.Table {
	t := tview.NewTable().SetBorders(false).SetSelectable(false, false)
	applyTheme(t)
	return t
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
