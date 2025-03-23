package tui

import (
	"context"
	"sync"
	"time"

	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

const (
	pageSummaryName    = "summary"
	pageSumDetailName  = "sumDetail"
	pageFlowDetailName = "flowDetail"
)

type flowAppState struct {
	SumID        int
	SumRow       int
	SumDetailRow int
	FlowID       int
}

type FlowApp struct {
	mu      *sync.Mutex
	app     *tview.Application
	fds     *flowdata.FlowDataStore
	fc      *flowcache.FlowCache
	fas     *flowAppState
	stopped bool
	pages   *tview.Pages
}

func NewFlowApp(fds *flowdata.FlowDataStore, fc *flowcache.FlowCache) *FlowApp {
	setTheme()
	pages := tview.NewPages()
	pages.SetBackgroundColor(bgColor).SetBorderColor(borderColor).SetTitleColor(titleColor)
	pages.SetBorderStyle(tcell.StyleDefault.Foreground(borderColor).Background(bgColor))
	fas := &flowAppState{}
	return &FlowApp{&sync.Mutex{}, tview.NewApplication(), fds, fc, fas, false, pages}
}

func (fa *FlowApp) Run(ctx context.Context) error {
	defer fa.Stop()
	cmdctx := cmdContext.CmdContextFromContext(ctx)

	// Set up an input capture to shutdown the app when the user presses Ctrl-C
	fa.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC || (event.Key() == tcell.KeyRune && event.Rune() == 'q') {
			fa.Stop()
			cmdctx.Cancel()
			return nil
		}
		if event.Key() == tcell.KeyRune && event.Rune() == '/' {
			return event
		}
		return event
	})

	// Go update screen periodically until we're shutdown
	go func() {
		ticker := time.Tick(2 * time.Second)
		for {
			select {
			case <-ctx.Done():
				log.Debugf("flowApp shutting down: done signal received")
				fa.Stop()
				return
			case <-ticker:
				fa.app.Draw()
			}
		}
	}()

	fa.pages.AddPage(pageSummaryName, fa.viewSummary(), true, true)
	fa.pages.AddPage(pageSumDetailName, fa.viewSumDetail(), true, false)

	// Start with a summary view
	if err := fa.app.SetRoot(fa.pages, true).Run(); err != nil {
		return err
	}
	return nil
}

func (fa *FlowApp) Stop() {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	if fa.app != nil && !fa.stopped {
		fa.app.Stop()
		fa.stopped = true
	}
}

func (fa *FlowApp) viewPage(pageName string) {
	fa.pages.SwitchToPage(pageName)
}

// func main() {
// 	app := tview.NewApplication()
//
// 	// Create the main layout
// 	mainLayout := tview.NewFlex().
// 		SetDirection(tview.FlexRow).
// 		AddItem(tview.NewTextView().
// 			SetTextAlign(tview.AlignCenter).
// 			SetText("Press 'Ctrl+F' to open the form dialog"), 0, 1, false)
//
// 	// Create the form that will go in the modal
// 	form := tview.NewForm().
// 		AddInputField("Name", "", 20, nil, nil).
// 		AddInputField("Email", "", 30, nil, nil).
// 		AddPasswordField("Password", "", 20, '*', nil).
// 		AddCheckbox("Send welcome email", false, nil).
// 		AddTextArea("Notes", "", 40, 4, 0, nil).
// 		AddButton("Save", func() {
// 			// Save form data here
// 			app.QueueUpdateDraw(func() {
// 				// Close the modal
// 				pages.SwitchToPage("main")
// 			})
// 		}).
// 		AddButton("Cancel", func() {
// 			app.QueueUpdateDraw(func() {
// 				// Close the modal without saving
// 				pages.SwitchToPage("main")
// 			})
// 		})
//
// 	// Style the form
// 	form.SetBorder(true).
// 		SetTitle("User Information").
// 		SetTitleAlign(tview.AlignCenter).
// 		SetBorderColor(tcell.ColorSteelBlue)
//
// 	// Create a flex container for the modal to center the form
// 	modalFlex := tview.NewFlex().
// 		SetDirection(tview.FlexRow).
// 		AddItem(nil, 0, 1, false).
// 		AddItem(tview.NewFlex().
// 			SetDirection(tview.FlexColumn).
// 			AddItem(nil, 0, 1, false).
// 			AddItem(form, 50, 1, true). // Width of 50
// 			AddItem(nil, 0, 1, false),
// 			10, 1, true). // Height of 10
// 		AddItem(nil, 0, 1, false)
//
// 	// Create modal frame with semi-transparent background
// 	modal := tview.NewFlex().
// 		SetBackgroundColor(tcell.ColorBlack.WithAlpha(192)) // Semi-transparent black
// 	modal.AddItem(modalFlex, 0, 1, true)
//
// 	// Create pages to switch between main and modal
// 	pages := tview.NewPages().
// 		AddPage("main", mainLayout, true, true).
// 		AddPage("modal", modal, true, false) // Modal starts hidden
//
// 	// Add keyboard shortcuts
// 	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
// 		if event.Key() == tcell.KeyCtrlF {
// 			// Show the modal when Ctrl+F is pressed
// 			pages.ShowPage("modal")
// 			return nil
// 		} else if event.Key() == tcell.KeyEscape {
// 			// Hide the modal when Escape is pressed
// 			if pages.HasPage("modal") {
// 				pages.HidePage("modal")
// 				return nil
// 			}
// 		}
// 		return event
// 	})
//
// 	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
// 		panic(err)
// 	}
// }
