package tui

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/rivo/tview"
)

// ButtonConfig represents configuration for a button
type ButtonConfig struct {
	Label    string
	Callback func()
}

// StreamViewer represents a dialog that can display command output or channel messages
type StreamViewer struct {
	app      *tview.Application
	frame    *tview.Frame
	textView *tview.TextView
	buttons  []*tview.Button
	flex     *tview.Flex
	onClose  func()
	mu       sync.Mutex
}

// NewStreamViewer creates a new stream viewer dialog with customizable buttons
func NewStreamViewer(app *tview.Application, buttonConfigs ...ButtonConfig) *StreamViewer {
	sv := &StreamViewer{
		app:      app,
		textView: tview.NewTextView(),
		buttons:  make([]*tview.Button, 0),
	}

	// Configure text view
	sv.textView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			// Auto-scroll to bottom when content changes
			sv.app.Draw()
			sv.textView.ScrollToEnd()
		})

	// If no buttons provided, add default OK button
	if len(buttonConfigs) == 0 {
		buttonConfigs = []ButtonConfig{
			{
				Label: "OK",
				Callback: func() {
					if sv.onClose != nil {
						sv.onClose()
					}
				},
			},
		}
	}

	// Create buttons
	buttonFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	// Add spacer at the beginning
	buttonFlex.AddItem(nil, 0, 1, false)

	for i, config := range buttonConfigs {
		btn := tview.NewButton(config.Label).
			SetSelectedFunc(config.Callback)
		sv.buttons = append(sv.buttons, btn)

		// Determine if this button should have focus initially (first button)
		shouldFocus := i == 0

		// Add button with fixed width
		buttonFlex.AddItem(btn, len(config.Label)+4, 0, shouldFocus)

		// Add small spacer between buttons (except after last button)
		if i < len(buttonConfigs)-1 {
			buttonFlex.AddItem(nil, 2, 0, false)
		}
	}

	// Add spacer at the end
	buttonFlex.AddItem(nil, 0, 1, false)

	// Create main flex layout
	sv.flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(sv.textView, 0, 1, false). // Text view takes most space
		AddItem(buttonFlex, 1, 0, true)    // Button row with fixed height

	// Create frame to add border and title
	sv.frame = tview.NewFrame(sv.flex).
		SetBorders(1, 1, 1, 1, 2, 2)

	sv.frame.SetBorder(true).
		SetTitle(" Stream Output ").
		SetTitleAlign(tview.AlignLeft)

	return sv
}

// NewStreamViewerWithOK creates a stream viewer with just an OK button (convenience function)
func NewStreamViewerWithOK(app *tview.Application) *StreamViewer {
	return NewStreamViewer(app)
}

// NewStreamViewerWithButtons creates a stream viewer with custom button labels
func NewStreamViewerWithButtons(app *tview.Application, labels ...string) *StreamViewer {
	configs := make([]ButtonConfig, len(labels))
	for i, label := range labels {
		labelCopy := label // Capture for closure
		configs[i] = ButtonConfig{
			Label: labelCopy,
			Callback: func() {
				// Default callback - can be overridden with SetButtonCallback
			},
		}
	}
	return NewStreamViewer(app, configs...)
}

// SetButtonCallback sets the callback for a button by index
func (sv *StreamViewer) SetButtonCallback(index int, callback func()) *StreamViewer {
	if index >= 0 && index < len(sv.buttons) {
		sv.buttons[index].SetSelectedFunc(callback)
	}
	return sv
}

// GetButton returns the button at the specified index
func (sv *StreamViewer) GetButton(index int) *tview.Button {
	if index >= 0 && index < len(sv.buttons) {
		return sv.buttons[index]
	}
	return nil
}

// GetButtonCount returns the number of buttons
func (sv *StreamViewer) GetButtonCount() int {
	return len(sv.buttons)
}

// SetTitle sets the title of the stream viewer dialog
func (sv *StreamViewer) SetTitle(title string) *StreamViewer {
	sv.frame.SetTitle(fmt.Sprintf(" %s ", title))
	return sv
}

// SetOnClose sets the callback function when dialog is closed
func (sv *StreamViewer) SetOnClose(fn func()) *StreamViewer {
	sv.onClose = fn
	return sv
}

// RunCommand executes a command and displays its stdout and stderr
func (sv *StreamViewer) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Get stderr pipe
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Read stdout
	go sv.readPipe(stdout, "[white]")

	// Read stderr
	go sv.readPipe(stderr, "[red]")

	// Wait for command to complete in background
	go func() {
		if err := cmd.Wait(); err != nil {
			sv.appendText(fmt.Sprintf("[yellow]Command exited with error: %v[-]\n", err))
		} else {
			sv.appendText("[green]Command completed successfully[-]\n")
		}
	}()

	return nil
}

// ReadFromChannel reads messages from a channel and displays them
func (sv *StreamViewer) ReadFromChannel(msgChan <-chan string) {
	go func() {
		for msg := range msgChan {
			sv.appendText(msg)
			if !strings.HasSuffix(msg, "\n") {
				sv.appendText("\n")
			}
		}
		sv.appendText("[green]Channel closed[-]\n")
	}()
}

// readPipe reads from an io.Reader and displays the content
func (sv *StreamViewer) readPipe(r io.Reader, colorTag string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		sv.appendText(fmt.Sprintf("%s%s[-]\n", colorTag, line))
	}
	if err := scanner.Err(); err != nil {
		sv.appendText(fmt.Sprintf("[red]Error reading: %v[-]\n", err))
	}
}

// appendText safely appends text to the text view
func (sv *StreamViewer) appendText(text string) {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	sv.app.QueueUpdateDraw(func() {
		currentText := sv.textView.GetText(false)
		sv.textView.SetText(currentText + text)
		sv.textView.ScrollToEnd()
	})
}

// AppendText publicly exposes text appending functionality
func (sv *StreamViewer) AppendText(text string) {
	sv.appendText(text)
}

// Clear clears the text view content
func (sv *StreamViewer) Clear() *StreamViewer {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	sv.textView.Clear()
	return sv
}

// GetPrimitive returns the primitive to be added to pages or displayed
func (sv *StreamViewer) GetPrimitive() tview.Primitive {
	return sv.frame
}

// Focus sets focus to the first button
func (sv *StreamViewer) Focus() {
	if len(sv.buttons) > 0 {
		sv.app.SetFocus(sv.buttons[0])
	}
}

// CreateCenteredViewer creates a centered modal-like appearance for the stream viewer
func CreateCenteredViewer(viewer *StreamViewer, width, height int) tview.Primitive {
	// Create a flex layout to center the viewer
	return tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(viewer.GetPrimitive(), height, 1, true).
				AddItem(nil, 0, 1, false),
			width, 1, true).
		AddItem(nil, 0, 1, false)
}

// // Example usage
// func main() {
// 	app := tview.NewApplication()
// 	pages := tview.NewPages()
//
// 	// Create main view with buttons to trigger stream viewers
// 	mainView := tview.NewFlex().
// 		SetDirection(tview.FlexRow)
//
// 	info := tview.NewTextView().
// 		SetText("Press buttons to test different stream viewer configurations:").
// 		SetTextAlign(tview.AlignCenter)
//
// 	buttonFlex := tview.NewFlex().
// 		SetDirection(tview.FlexColumn)
//
// 	// Example 1: Default OK button
// 	defaultButton := tview.NewButton("Default (OK)").
// 		SetSelectedFunc(func() {
// 			viewer := NewStreamViewerWithOK(app).
// 				SetTitle("Default OK Button").
// 				SetOnClose(func() {
// 					pages.RemovePage("viewer")
// 					pages.SwitchToPage("main")
// 				})
//
// 			viewer.AppendText("[cyan]This viewer has the default OK button[-]\n")
// 			viewer.AppendText("Click OK to close.\n")
//
// 			centeredViewer := CreateCenteredViewer(viewer, 60, 15)
// 			pages.AddPage("viewer", centeredViewer, true, true)
// 			viewer.Focus()
// 		})
//
// 	// Example 2: Custom buttons with callbacks
// 	customButton := tview.NewButton("Custom Buttons").
// 		SetSelectedFunc(func() {
// 			var viewer *StreamViewer
// 			viewer = NewStreamViewer(app,
// 				ButtonConfig{
// 					Label: "Clear",
// 					Callback: func() {
// 						viewer.Clear()
// 						viewer.AppendText("[green]Content cleared![-]\n")
// 					},
// 				},
// 				ButtonConfig{
// 					Label: "Add Line",
// 					Callback: func() {
// 						viewer.AppendText("[yellow]New line added![-]\n")
// 					},
// 				},
// 				ButtonConfig{
// 					Label: "Close",
// 					Callback: func() {
// 						pages.RemovePage("viewer")
// 						pages.SwitchToPage("main")
// 					},
// 				},
// 			).SetTitle("Custom Buttons Example")
//
// 			viewer.AppendText("[cyan]This viewer has custom buttons[-]\n")
// 			viewer.AppendText("Try clicking the different buttons!\n")
//
// 			centeredViewer := CreateCenteredViewer(viewer, 70, 20)
// 			pages.AddPage("viewer", centeredViewer, true, true)
// 			viewer.Focus()
// 		})
//
// 	// Example 3: Yes/No dialog style
// 	yesNoButton := tview.NewButton("Yes/No Style").
// 		SetSelectedFunc(func() {
// 			viewer := NewStreamViewerWithButtons(app, "Yes", "No", "Cancel").
// 				SetTitle("Confirmation Dialog")
//
// 			// Set callbacks for each button
// 			viewer.SetButtonCallback(0, func() { // Yes
// 				pages.RemovePage("viewer")
// 				pages.SwitchToPage("main")
// 			}).SetButtonCallback(1, func() { // No
// 				pages.RemovePage("viewer")
// 				pages.SwitchToPage("main")
// 			}).SetButtonCallback(2, func() { // Cancel
// 				pages.RemovePage("viewer")
// 				pages.SwitchToPage("main")
// 			})
//
// 			viewer.AppendText("[white]Do you want to proceed?[-]\n\n")
// 			viewer.AppendText("This is an example of a Yes/No/Cancel dialog.\n")
//
// 			centeredViewer := CreateCenteredViewer(viewer, 50, 12)
// 			pages.AddPage("viewer", centeredViewer, true, true)
// 			viewer.Focus()
// 		})
//
// 	// Example 4: Command with Retry
// 	commandButton := tview.NewButton("Command + Retry").
// 		SetSelectedFunc(func() {
// 			var viewer *StreamViewer
//
// 			runCommand := func() {
// 				viewer.Clear()
// 				viewer.AppendText("[cyan]Running command...[-]\n\n")
// 				if err := viewer.RunCommand("ls", "-la"); err != nil {
// 					viewer.AppendText(fmt.Sprintf("[red]Error: %v[-]\n", err))
// 				}
// 			}
//
// 			viewer = NewStreamViewer(app,
// 				ButtonConfig{
// 					Label:    "Retry",
// 					Callback: runCommand,
// 				},
// 				ButtonConfig{
// 					Label: "OK",
// 					Callback: func() {
// 						pages.RemovePage("viewer")
// 						pages.SwitchToPage("main")
// 					},
// 				},
// 			).SetTitle("Command with Retry")
//
// 			runCommand()
//
// 			pages.AddPage("viewer", viewer.GetPrimitive(), true, true)
// 			viewer.Focus()
// 		})
//
// 	// Quit button
// 	quitButton := tview.NewButton("Quit").
// 		SetSelectedFunc(func() {
// 			app.Stop()
// 		})
//
// 	buttonFlex.
// 		AddItem(defaultButton, 0, 1, true).
// 		AddItem(customButton, 0, 1, false).
// 		AddItem(yesNoButton, 0, 1, false).
// 		AddItem(commandButton, 0, 1, false).
// 		AddItem(quitButton, 0, 1, false)
//
// 	mainView.
// 		AddItem(info, 3, 0, false).
// 		AddItem(buttonFlex, 3, 0, true).
// 		AddItem(tview.NewTextView(), 0, 1, false) // Spacer
//
// 	mainView.SetBorder(true).SetTitle(" Stream Viewer Demo ")
//
// 	pages.AddPage("main", mainView, true, true)
//
// 	// Set up input capture for navigation
// 	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
// 		// Check if viewer is shown
// 		if pages.HasPage("viewer") {
// 			if event.Key() == tcell.KeyEscape {
// 				pages.RemovePage("viewer")
// 				pages.SwitchToPage("main")
// 				return nil
// 			}
// 			return event
// 		}
//
// 		// Tab navigation between buttons on main page
// 		if event.Key() == tcell.KeyTab {
// 			if defaultButton.HasFocus() {
// 				app.SetFocus(customButton)
// 			} else if customButton.HasFocus() {
// 				app.SetFocus(yesNoButton)
// 			} else if yesNoButton.HasFocus() {
// 				app.SetFocus(commandButton)
// 			} else if commandButton.HasFocus() {
// 				app.SetFocus(quitButton)
// 			} else {
// 				app.SetFocus(defaultButton)
// 			}
// 			return nil
// 		}
// 		return event
// 	})
//
// 	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
// 		panic(err)
// 	}
// }
