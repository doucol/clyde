package catcher

import (
	"bufio"
	"context"
	"errors"
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

type pfOut struct{}

func (l pfOut) Write(bytes []byte) (int, error) {
	log.Debugf("portforward stdout: %s", string(bytes))
	return len(bytes), nil
}

type pfErr struct{}

func (l pfErr) Write(bytes []byte) (int, error) {
	log.Debugf("portforward stderr: %s", string(bytes))
	return len(bytes), nil
}

func (dc *DataCatcher) CatchDataFromSSEStream() error {
	select {
	case <-dc.ctx.Done():
		log.Debug("done signal received - not entering CatchDataFromSSEStream")
		return nil
	default:
	}
	log.Debug("entering flow catcher")

	// Channels for signaling
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})

	shutdown := false
	go func() {
		// Shutdown the port forwarding when the context is done
		// This will force the ConsumeSSEStream to exit below
		<-dc.ctx.Done()
		log.Debug("done signal received - sending signal to shutdown port foward")
		shutdown = true
		stopChan <- struct{}{}
	}()

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

	defer func() {
		// Shutdown the port forwarding in case we've exited for some other reason
		select {
		case stopChan <- struct{}{}:
		default:
		}
		log.Debug("exiting flow catcher")
	}()

	pf, err := portforward.New(dialer, ports, stopChan, readyChan, pfOut{}, pfErr{})
	if err != nil {
		return err
	}

	// Start the port forwarding
	go func() {
		log.Debugf("Starting port forward from localhost:%d to %s/%s:%s", freePort, dc.namespace, podName, port)
		if err := pf.ForwardPorts(); err != nil && !shutdown {
			log.Debugf("error: ForwardPorts return error: %s", err.Error())
		}
		log.Debug("port forward has stopped")
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
			log.Debug("done signal received: in split func - sending back EOF")
			return 0, nil, io.EOF
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
			log.Debugf("SSE stream data discarded: %s", line)
			return nil
		}
	}

	// Handle any errors during scanning
	err = scanner.Err()
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Debug("EOF received from scanner in ConsumeSSEStream, exiting now")
		} else {
			return err
		}
	}
	return nil
}
