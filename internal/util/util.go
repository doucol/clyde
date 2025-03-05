package util

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/doucol/clyde/internal/cmdContext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
)

func GetPodAndEnvVarWithContainerName(ctx context.Context, namespace string, containerName string, envVarName string) (string, string, error) {
	clientset := cmdContext.ClientsetFromContext(ctx)
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", "", err
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if container.Name == containerName {
				for _, env := range container.Env {
					if env.Name == envVarName {
						return pod.Name, env.Value, nil
					}
				}
			}
		}
	}

	return "", "", fmt.Errorf("pod or env var name not found")
}

// ConsumeSSEStream connects to an SSE endpoint and processes events.
func ConsumeSSEStream(ctx context.Context, url string, cb func(data string)) error {
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
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		select {
		case <-ctx.Done():
			return 0, nil, context.Canceled
		default:
			return bufio.ScanLines(data, atEOF)
		}
	})

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
			cb(data)
		}
	}

	// Handle any errors during scanning
	return scanner.Err()
}

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func GetDataPath() string {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		dataDir = filepath.Join(homedir.HomeDir(), ".local", "share")
	}
	dataDir = filepath.Join(dataDir, "clyde")
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		panic(err)
	}
	return dataDir
}
