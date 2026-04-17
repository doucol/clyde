package tui

import "charm.land/bubbles/v2/key"

type keyMap struct {
	Quit        key.Binding
	Back        key.Binding
	Help        key.Binding
	Filter      key.Binding
	Home        key.Binding
	Rates       key.Binding
	Totals      key.Binding
	SortKey     key.Binding
	SortSrcPkt  key.Binding
	SortDstPkt  key.Binding
	SortSrcByte key.Binding
	SortDstByte key.Binding
	Up          key.Binding
	Down        key.Binding
	PageUp      key.Binding
	PageDown    key.Binding
	GotoTop     key.Binding
	GotoBottom  key.Binding
	Enter       key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Home: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "home"),
		),
		Rates: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rates view"),
		),
		Totals: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "totals view"),
		),
		SortKey: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "sort by key"),
		),
		SortSrcPkt: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "sort src pkt rate"),
		),
		SortDstPkt: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "sort dst pkt rate"),
		),
		SortSrcByte: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "sort src byte rate"),
		),
		SortDstByte: key.NewBinding(
			key.WithKeys("B"),
			key.WithHelp("B", "sort dst byte rate"),
		),
		Up:         key.NewBinding(key.WithKeys("up", "k")),
		Down:       key.NewBinding(key.WithKeys("down", "j")),
		PageUp:     key.NewBinding(key.WithKeys("pgup")),
		PageDown:   key.NewBinding(key.WithKeys("pgdown")),
		GotoTop:    key.NewBinding(key.WithKeys("home", "g")),
		GotoBottom: key.NewBinding(key.WithKeys("end", "G")),
		Enter:      key.NewBinding(key.WithKeys("enter")),
	}
}

var keys = newKeyMap()

var helpEntries = []struct {
	Key, Description string
}{
	{"q or Ctrl+C", "Quit the application"},
	{"h", "Return to the context picker"},
	{"r", "Switch to Rates view (from Totals)"},
	{"t", "Switch to Totals view (from Rates)"},
	{"p", "Sort by Source Packet Rate (rates only)"},
	{"P", "Sort by Dest Packet Rate (rates only)"},
	{"b", "Sort by Source Byte Rate (rates only)"},
	{"B", "Sort by Dest Byte Rate (rates only)"},
	{"n", "Sort by Key (totals or rates)"},
	{"/", "Open filter dialog"},
	{"?", "Show this help dialog"},
}
