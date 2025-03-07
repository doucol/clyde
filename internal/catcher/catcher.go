package catcher

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type CatcherFunc func(data string) error

type DataCatcher struct {
	ctx            context.Context
	catcher        CatcherFunc
	namespace      string
	containerName  string
	portEnvVarName string
	urlPath        string
}

func NewDataCatcher(ctx context.Context, namespace, containerName, portEnvVarName, urlPath string, catcher CatcherFunc) *DataCatcher {
	return &DataCatcher{
		ctx:            ctx,
		namespace:      namespace,
		containerName:  containerName,
		portEnvVarName: portEnvVarName,
		urlPath:        urlPath,
		catcher:        catcher,
	}
}

func (dc *DataCatcher) CatchDataFromSSEStream() error {
	config := cmdContext.K8sConfigFromContext(dc.ctx)

	// URL for the portforward endpoint on the pod
	podName, port, err := util.GetPodAndEnvVarWithContainerName(dc.ctx, dc.namespace, dc.containerName, dc.portEnvVarName)
	if err != nil {
		return err
	}

	apiURL, _ := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", config.Host, dc.namespace, podName))

	// Dialer for establishing the connection
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiURL)

	// Port mappings (local port: pod port)
	freePort, err := util.GetFreePort()
	if err != nil {
		return err
	}
	ports := []string{fmt.Sprintf("%d:%s", freePort, port)}

	// Channels for signaling
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})

	shutdown := false
	go func() {
		// Shutdown the port forwarding when the context is done
		// This will force the ConsumeSSEStream to exit below
		<-dc.ctx.Done()
		shutdown = true
		stopChan <- struct{}{}
	}()

	defer func() {
		// Shutdown the port forwarding in case we've exited for some other reason
		select {
		case stopChan <- struct{}{}:
		default:
		}
	}()

	pf, err := portforward.New(dialer, ports, stopChan, readyChan, io.Discard, io.Discard)
	if err != nil {
		return err
	}

	// Start the port forwarding
	go func() {
		if err := pf.ForwardPorts(); err != nil && !shutdown {
			panic(err)
		}
	}()

	// Wait for the port forwarding to be ready
	<-readyChan

	sseURL := fmt.Sprintf("http://localhost:%d%s", freePort, dc.urlPath)
	if err := dc.ConsumeSSEStream(sseURL); err != nil && !shutdown {
		return err
	}

	return nil
}

// ConsumeSSEStream connects to an SSE endpoint and processes events.
func (dc *DataCatcher) ConsumeSSEStream(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to SSE stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		select {
		case <-dc.ctx.Done():
			return 0, nil, context.Canceled
		default:
			return bufio.ScanLines(data, atEOF)
		}
	})

	// Read the SSE stream line by line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Parse the event data
		if strings.HasPrefix(line, "data:") {
			log.Tracef("Stream data received: %s", line)
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			if err := dc.catcher(data); err != nil {
				return err
			}
		} else if line != "" {
			log.Warnf("SSE stream data discarded: %s", line)
		}
	}

	// Handle any errors during scanning
	return scanner.Err()
}
