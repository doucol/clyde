// Package tui provides a command dialog for display real-time command output or event channel messages
package tui

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

const commandDialogName = "commandDialog"

// CommandDialog represents a modal dialog for executing commands and displaying real-time output
type CommandDialog struct {
	modal       *tview.Modal
	textView    *tview.TextView
	flex        *tview.Flex
	pages       *tview.Pages
	command     string
	args        []string
	eventChan   <-chan string
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	isRunning   bool
	outputLines []string
	maxLines    int
	mode        string // "command" or "channel"
}

// CommandDialogConfig holds configuration for the command dialog
type CommandDialogConfig struct {
	Title     string
	Command   string
	Args      []string
	EventChan <-chan string // Optional: channel that emits strings for real-time display
	MaxLines  int           // Maximum number of output lines to keep (0 = unlimited)
	Width     int           // Dialog width (0 = auto)
	Height    int           // Dialog height (0 = auto)
}

// NewCommandDialog creates a new command dialog instance
func NewCommandDialog(pages *tview.Pages, config CommandDialogConfig) *CommandDialog {
	if config.MaxLines == 0 {
		config.MaxLines = 1000 // Default max lines
	}
	if config.Width == 0 {
		config.Width = 80 // Default width
	}
	if config.Height == 0 {
		config.Height = 24 // Default height
	}

	cd := &CommandDialog{
		pages:       pages,
		command:     config.Command,
		args:        config.Args,
		eventChan:   config.EventChan,
		maxLines:    config.MaxLines,
		outputLines: make([]string, 0),
	}

	// Determine mode based on configuration
	if config.EventChan != nil {
		cd.mode = "channel"
	} else {
		cd.mode = "command"
	}

	// Create the text view for output
	cd.textView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true)
	cd.textView.SetBorder(true).SetTitle("Output")
	cd.textView.SetBackgroundColor(bgColor).SetBorderColor(borderColor).SetTitleColor(titleColor)

	// Create buttons based on mode
	var executeButton *tview.Button
	if cd.mode == "command" {
		executeButton = tview.NewButton("Execute").SetSelectedFunc(func() {
			cd.executeCommand()
		})
	} else {
		executeButton = tview.NewButton("Start").SetSelectedFunc(func() {
			cd.startChannelListener()
		})
	}
	executeButton.SetBackgroundColor(bgColor)

	stopButton := tview.NewButton("Stop").SetSelectedFunc(func() {
		if cd.mode == "command" {
			cd.stopCommand()
		} else {
			cd.stopChannelListener()
		}
	})
	stopButton.SetBackgroundColor(bgColor)

	clearButton := tview.NewButton("Clear").SetSelectedFunc(func() {
		cd.clearOutput()
	})
	clearButton.SetBackgroundColor(bgColor)

	closeButton := tview.NewButton("Close").SetSelectedFunc(func() {
		cd.Close()
	})
	closeButton.SetBackgroundColor(bgColor)

	// Create button flex
	buttonFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(executeButton, 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(stopButton, 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(clearButton, 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(closeButton, 0, 1, false)

	// Create info display
	var infoText string
	if cd.mode == "command" {
		infoText = fmt.Sprintf("Command: %s %s", config.Command, strings.Join(config.Args, " "))
	} else {
		infoText = "Event Channel Listener"
	}
	commandInfo := tview.NewTextView().
		SetText(infoText).
		SetTextAlign(tview.AlignCenter)
	commandInfo.SetBackgroundColor(bgColor)

	// Create main flex container
	cd.flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(commandInfo, 1, 0, false).
		AddItem(cd.textView, 0, 1, false).
		AddItem(buttonFlex, 3, 0, false)

	cd.flex.SetBorder(true).SetTitle(config.Title)
	cd.flex.SetBackgroundColor(bgColor).SetBorderColor(borderColor).SetTitleColor(titleColor)

	// Set up input capture for keyboard shortcuts
	cd.flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			cd.Close()
			return nil
		case tcell.KeyCtrlC:
			cd.stopCommand()
			return nil
		}
		switch event.Rune() {
		case 'e', 'E':
			if cd.mode == "command" {
				cd.executeCommand()
			} else {
				cd.startChannelListener()
			}
			return nil
		case 's', 'S':
			if cd.mode == "command" {
				cd.stopCommand()
			} else {
				cd.stopChannelListener()
			}
			return nil
		case 'c', 'C':
			cd.clearOutput()
			return nil
		case 'q', 'Q':
			cd.Close()
			return nil
		}
		return event
	})

	return cd
}

// Show displays the command dialog
func (cd *CommandDialog) Show() {
	// Center the dialog
	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(cd.flex, 80, 0, true).
			AddItem(nil, 0, 1, false), 24, 0, true).
		AddItem(nil, 0, 1, false)

	cd.pages.AddPage(commandDialogName, modal, true, true)
}

// Close closes the command dialog
func (cd *CommandDialog) Close() {
	if cd.mode == "command" {
		cd.stopCommand()
	} else {
		cd.stopChannelListener()
	}
	cd.pages.RemovePage(commandDialogName)
}

// executeCommand starts executing the command
func (cd *CommandDialog) executeCommand() {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if cd.isRunning {
		cd.appendOutput("[yellow]Command is already running[white]")
		return
	}

	cd.isRunning = true
	cd.ctx, cd.cancel = context.WithCancel(context.Background())

	cd.appendOutput(fmt.Sprintf("[green]Starting command: %s %s[white]", cd.command, strings.Join(cd.args, " ")))

	go cd.runCommand()
}

// startChannelListener starts listening to the event channel
func (cd *CommandDialog) startChannelListener() {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if cd.isRunning {
		cd.appendOutput("[yellow]Channel listener is already running[white]")
		return
	}

	if cd.eventChan == nil {
		cd.appendOutput("[red]No event channel configured[white]")
		return
	}

	cd.isRunning = true
	cd.ctx, cd.cancel = context.WithCancel(context.Background())

	cd.appendOutput("[green]Starting channel listener[white]")

	go cd.runChannelListener()
}

// stopCommand stops the currently running command
func (cd *CommandDialog) stopCommand() {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if !cd.isRunning {
		return
	}

	if cd.cancel != nil {
		cd.cancel()
	}
	cd.isRunning = false
	cd.appendOutput("[red]Command stopped[white]")
}

// stopChannelListener stops the channel listener
func (cd *CommandDialog) stopChannelListener() {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if !cd.isRunning {
		return
	}

	if cd.cancel != nil {
		cd.cancel()
	}
	cd.isRunning = false
	cd.appendOutput("[red]Channel listener stopped[white]")
}

// clearOutput clears the output text view
func (cd *CommandDialog) clearOutput() {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	cd.outputLines = make([]string, 0)
	cd.textView.SetText("")
}

// runCommand executes the command and streams output
func (cd *CommandDialog) runCommand() {
	defer func() {
		cd.mu.Lock()
		cd.isRunning = false
		cd.mu.Unlock()
	}()

	cmd := exec.CommandContext(cd.ctx, cd.command, cd.args...)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cd.appendOutput(fmt.Sprintf("[red]Error creating stdout pipe: %v[white]", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cd.appendOutput(fmt.Sprintf("[red]Error creating stderr pipe: %v[white]", err))
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		cd.appendOutput(fmt.Sprintf("[red]Error starting command: %v[white]", err))
		return
	}

	// Create channels for output
	outputChan := make(chan string, 100)
	var wg sync.WaitGroup

	// Read stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case outputChan <- scanner.Text():
			case <-cd.ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			logrus.Errorf("Error reading stdout: %v", err)
		}
	}()

	// Read stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			select {
			case outputChan <- fmt.Sprintf("[red]%s[white]", scanner.Text()):
			case <-cd.ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			logrus.Errorf("Error reading stderr: %v", err)
		}
	}()

	// Close output channel when readers are done
	go func() {
		wg.Wait()
		close(outputChan)
	}()

	// Process output
	for {
		select {
		case line, ok := <-outputChan:
			if !ok {
				// Channel closed, command finished
				if err := cmd.Wait(); err != nil {
					cd.appendOutput(fmt.Sprintf("[red]Command finished with error: %v[white]", err))
				} else {
					cd.appendOutput("[green]Command completed successfully[white]")
				}
				return
			}
			cd.appendOutput(line)
		case <-cd.ctx.Done():
			// Command was cancelled
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			return
		}
	}
}

// runChannelListener listens to the event channel and displays messages
func (cd *CommandDialog) runChannelListener() {
	defer func() {
		cd.mu.Lock()
		cd.isRunning = false
		cd.mu.Unlock()
	}()

	for {
		select {
		case message, ok := <-cd.eventChan:
			if !ok {
				// Channel closed
				cd.appendOutput("[yellow]Event channel closed[white]")
				return
			}
			cd.appendOutput(message)
		case <-cd.ctx.Done():
			// Context cancelled
			return
		}
	}
}

// appendOutput adds a line to the output and updates the text view
func (cd *CommandDialog) appendOutput(line string) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	cd.outputLines = append(cd.outputLines, line)

	// Trim lines if we exceed max
	if len(cd.outputLines) > cd.maxLines {
		cd.outputLines = cd.outputLines[len(cd.outputLines)-cd.maxLines:]
	}

	// Update the text view
	text := strings.Join(cd.outputLines, "\n")
	cd.textView.SetText(text)
	cd.textView.ScrollToEnd()
}

// ExecuteCommandInDialog is a convenience function to show a command dialog for the FlowApp
func (fa *FlowApp) ExecuteCommandInDialog(config CommandDialogConfig) {
	dialog := NewCommandDialog(fa.pages, config)
	dialog.Show()
}
