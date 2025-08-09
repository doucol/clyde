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

// CommandModal represents a modal dialog that can display command output or channel messages
type CommandModal struct {
	app      *tview.Application
	modal    *tview.Modal
	textView *tview.TextView
	flex     *tview.Flex
	onClose  func()
	mu       sync.Mutex
}

// NewCommandModal creates a new command modal dialog
func NewCommandModal(app *tview.Application) *CommandModal {
	cm := &CommandModal{
		app:      app,
		textView: tview.NewTextView(),
	}

	// Configure text view
	cm.textView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			// Auto-scroll to bottom when content changes
			cm.app.Draw()
			cm.textView.ScrollToEnd()
		})

	// Create modal with only OK button
	cm.modal = tview.NewModal().
		SetText("").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if cm.onClose != nil {
				cm.onClose()
			}
		})

	// Create flex layout
	cm.flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cm.textView, 0, 1, false).
		AddItem(cm.modal, 3, 0, true)

	// Set border and title
	cm.flex.SetBorder(true).
		SetTitle(" Command Output ").
		SetTitleAlign(tview.AlignLeft)

	return cm
}

// SetTitle sets the title of the modal dialog
func (cm *CommandModal) SetTitle(title string) *CommandModal {
	cm.flex.SetTitle(fmt.Sprintf(" %s ", title))
	return cm
}

// SetOnClose sets the callback function when modal is closed
func (cm *CommandModal) SetOnClose(fn func()) *CommandModal {
	cm.onClose = fn
	return cm
}

// RunCommand executes a command and displays its stdout and stderr
func (cm *CommandModal) RunCommand(name string, args ...string) error {
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
	go cm.readPipe(stdout, "[white]")

	// Read stderr
	go cm.readPipe(stderr, "[red]")

	// Wait for command to complete in background
	go func() {
		if err := cmd.Wait(); err != nil {
			cm.appendText(fmt.Sprintf("[yellow]Command exited with error: %v[-]\n", err))
		} else {
			cm.appendText("[green]Command completed successfully[-]\n")
		}
	}()

	return nil
}

// ReadFromChannel reads messages from a channel and displays them
func (cm *CommandModal) ReadFromChannel(msgChan <-chan string) {
	go func() {
		for msg := range msgChan {
			cm.appendText(msg)
			if !strings.HasSuffix(msg, "\n") {
				cm.appendText("\n")
			}
		}
		cm.appendText("[green]Channel closed[-]\n")
	}()
}

// readPipe reads from an io.Reader and displays the content
func (cm *CommandModal) readPipe(r io.Reader, colorTag string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		cm.appendText(fmt.Sprintf("%s%s[-]\n", colorTag, line))
	}
	if err := scanner.Err(); err != nil {
		cm.appendText(fmt.Sprintf("[red]Error reading: %v[-]\n", err))
	}
}

// appendText safely appends text to the text view
func (cm *CommandModal) appendText(text string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.app.QueueUpdateDraw(func() {
		currentText := cm.textView.GetText(false)
		cm.textView.SetText(currentText + text)
		cm.textView.ScrollToEnd()
	})
}

// Clear clears the text view content
func (cm *CommandModal) Clear() *CommandModal {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.textView.Clear()
	return cm
}

// GetPrimitive returns the primitive to be added to pages or displayed
func (cm *CommandModal) GetPrimitive() tview.Primitive {
	return cm.flex
}

// // Example usage
// func main() {
// 	app := tview.NewApplication()
// 	pages := tview.NewPages()
//
// 	// Create main view with buttons to trigger modals
// 	mainView := tview.NewFlex().
// 		SetDirection(tview.FlexRow)
//
// 	info := tview.NewTextView().
// 		SetText("Press buttons to test the command modal:").
// 		SetTextAlign(tview.AlignCenter)
//
// 	buttonFlex := tview.NewFlex().
// 		SetDirection(tview.FlexColumn)
//
// 	// Button to run a command
// 	cmdButton := tview.NewButton("Run Command").
// 		SetSelectedFunc(func() {
// 			modal := NewCommandModal(app).
// 				SetTitle("Running ls -la").
// 				SetOnClose(func() {
// 					pages.RemovePage("modal")
// 				})
//
// 			// Example: Run ls command
// 			if err := modal.RunCommand("ls", "-la"); err != nil {
// 				modal.appendText(fmt.Sprintf("[red]Error: %v[-]\n", err))
// 			}
//
// 			pages.AddPage("modal", modal.GetPrimitive(), true, true)
// 		})
//
// 	// Button to test channel messages
// 	channelButton := tview.NewButton("Test Channel").
// 		SetSelectedFunc(func() {
// 			modal := NewCommandModal(app).
// 				SetTitle("Channel Messages").
// 				SetOnClose(func() {
// 					pages.RemovePage("modal")
// 				})
//
// 			// Create a channel and send messages
// 			msgChan := make(chan string)
// 			modal.ReadFromChannel(msgChan)
//
// 			// Send test messages
// 			go func() {
// 				for i := 1; i <= 10; i++ {
// 					msgChan <- fmt.Sprintf("[cyan]Message %d[-]: This is a test message", i)
// 					// Small delay to see real-time updates
// 					// In real usage, messages would come from your actual source
// 				}
// 				close(msgChan)
// 			}()
//
// 			pages.AddPage("modal", modal.GetPrimitive(), true, true)
// 		})
//
// 	// Button to quit
// 	quitButton := tview.NewButton("Quit").
// 		SetSelectedFunc(func() {
// 			app.Stop()
// 		})
//
// 	buttonFlex.
// 		AddItem(cmdButton, 0, 1, true).
// 		AddItem(channelButton, 0, 1, false).
// 		AddItem(quitButton, 0, 1, false)
//
// 	mainView.
// 		AddItem(info, 3, 0, false).
// 		AddItem(buttonFlex, 3, 0, true).
// 		AddItem(tview.NewTextView(), 0, 1, false) // Spacer
//
// 	mainView.SetBorder(true).SetTitle(" Command Modal Demo ")
//
// 	pages.AddPage("main", mainView, true, true)
//
// 	// Set up input capture for navigation
// 	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
// 		if event.Key() == tcell.KeyTab {
// 			// Tab navigation between buttons
// 			if cmdButton.HasFocus() {
// 				app.SetFocus(channelButton)
// 			} else if channelButton.HasFocus() {
// 				app.SetFocus(quitButton)
// 			} else {
// 				app.SetFocus(cmdButton)
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
