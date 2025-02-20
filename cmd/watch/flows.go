package watch

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// ConsumeSSEStream connects to an SSE endpoint and processes events.
func ConsumeSSEStream(url string) error {
	// Create an HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to SSE stream: %w", err)
	}
	defer resp.Body.Close()

	// Check for a valid response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the SSE stream line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments or empty lines
		if strings.HasPrefix(line, ":") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Parse the event data
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			fmt.Printf("Received event data: %s\n", data)
		}
	}

	// Handle any errors during scanning
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading SSE stream: %w", err)
	}

	return nil
}

var WatchFlowsCmd = &cobra.Command{
	Use:   "flows",
	Short: "Watch calico flows",
	Long:  `Watch live calico flows in near real-time`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sseURL := "http://localhost:3002/flows/_stream"
		if err := ConsumeSSEStream(sseURL); err != nil {
			return fmt.Errorf("error consuming SSE stream: %w", err)
		}
		return nil
	},
}

// fmt.Printf("watch a live view of allow/deny flows...coming soon\n")
// return fmt.Errorf("not yet implemented")
// clientset := internal.ClientsetFromContext(cmd.Context())
// // pods, err := clientset.CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{})
// // if err != nil {
// // 	return err
// // }
// watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
// 	timeOut := int64(60)
// 	return clientset.CoreV1().Pods("kube-system").Watch(context.Background(), metav1.ListOptions{TimeoutSeconds: &timeOut})
// }
//
// watcher, _ := toolsWatch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})
//
// app := tview.NewApplication()
// table := tview.NewTable().SetBorders(false).SetSelectable(true, false).SetContent()
// table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
// 	if key == tcell.KeyEscape {
// 		app.Stop()
// 	}
// 	if key == tcell.KeyEnter {
// 		table.SetSelectable(true, true)
// 	}
// }).SetSelectedFunc(func(row int, column int) {
// 	table.GetCell(row, column).SetTextColor(tcell.ColorRed)
// 	table.SetSelectable(false, false)
// })
// if err := app.SetRoot(table, true).SetFocus(table).Run(); err != nil {
// 	return err
// }
//
// items := map[string]gridPod{}
//
// for event := range watcher.ResultChan() {
// 	item := event.Object.(*corev1.Pod)
//
// 	switch event.Type {
// 	case watch.Modified:
// 		break
// 	case watch.Bookmark:
// 		break
// 	case watch.Error:
// 		break
// 	case watch.Deleted:
// 		delete(items, item.GetName())
// 		break
// 	case watch.Added:
// 		name := item.GetName()
// 		cell := tview.NewTableCell(name).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignCenter)
// 		items[name] = gridPod{}
// 		table.SetCell(row, column, cell)
// 		// fmt.Printf("pod watch trigger on %v: Name: %v\n", event.Type, item.GetName())
// 	}
// }

// for row, pod := range pods.Items {
// 	// fmt.Printf("Pod name: %s\n", pod.Name)
// 	column := 0
// 	color := tcell.ColorWhite
// 	if column < 1 || row < 1 {
// 		color = tcell.ColorYellow
// 	}
// 	apiVersion, kind := pod.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
// 	cell1 := tview.NewTableCell(kind).SetTextColor(color).SetAlign(tview.AlignCenter)
// 	table.SetCell(row, column, cell1)
// 	// fmt.Fprintf(os.Stdout, "pod: %v\n\n", pod)
// 	column += 1
// 	cell2 := tview.NewTableCell(apiVersion).SetTextColor(color).SetAlign(tview.AlignCenter)
// 	table.SetCell(row, column, cell2)
// 	column += 1
// 	cell3 := tview.NewTableCell(pod.Name).SetTextColor(color).SetAlign(tview.AlignCenter)
// 	table.SetCell(row, column, cell3)
// 	column += 1
// }

// return nil
