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
	"sync"
	"time"

	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

var ErrUnknownSSEData = errors.New("unknown data format in SSE stream")

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

type CatcherFunc func(data string) error

type DataCatcher struct {
	catcher        CatcherFunc
	namespace      string
	containerName  string
	urlPath        string
	PortEnvVarName string
}

func NewDataCatcher(namespace, containerName, urlPath string, catcher CatcherFunc) *DataCatcher {
	return &DataCatcher{
		namespace:      namespace,
		containerName:  containerName,
		urlPath:        urlPath,
		catcher:        catcher,
		PortEnvVarName: "PORT",
	}
}

func (dc *DataCatcher) CatchDataFromSSEStream(ctx context.Context) error {
	select {
	case <-ctx.Done():
		log.Debug("done signal received - not entering CatchDataFromSSEStream")
		return nil
	default:
		log.Debug("entering data catcher")
	}

	config := cmdContext.K8sConfigFromContext(ctx)

	// URL for the portforward endpoint on the pod
	podName, port, err := util.GetPodAndEnvVarByContainerName(ctx, dc.namespace, dc.containerName, dc.PortEnvVarName)
	if err != nil {
		return err
	}

	apiURL, _ := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", config.Host, dc.namespace, podName))

	// Dialer for establishing the connection
	log.Debugf("apiURL: %s", apiURL)
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
	stopChan := make(chan struct{}, 20)
	readyChan := make(chan struct{})

	defer func() {
		util.ChanClose(stopChan, readyChan)
		log.Debug("exited data catcher")
	}()

	pf, err := portforward.New(dialer, ports, stopChan, readyChan, pfOut{}, pfErr{})
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}

	// Start the port forwarding
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer util.ChanSendEmpty(stopChan, 2)
		log.Debugf("Starting port forward from localhost:%d to %s/%s:%s", freePort, dc.namespace, podName, port)
		if err := pf.ForwardPorts(); err != nil {
			log.Debugf("error: ForwardPorts return error: %s", err.Error())
		}
		log.Debug("port forward has stopped")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer util.ChanSendEmpty(stopChan, 2)
		select {
		case <-readyChan:
			// Wait for the port forwarding to be ready
			sseURL := fmt.Sprintf("http://localhost:%d%s", freePort, dc.urlPath)
			if err := dc.consumeSSEStream(ctx, sseURL, stopChan); err != nil {
				log.Debugf("error: ConsumeSSEStream return error: %s", err.Error())
			}
		case <-time.After(5 * time.Second):
			log.Debug("timeout waiting for port forward to be ready")
		}
		log.Debug("sse consumer has stopped")
	}()

	select {
	case <-ctx.Done():
		util.ChanSendEmpty(stopChan, 2)
		log.Debug("done signal received, now waiting for port forward and sse streamer to exit")
	case <-stopChan:
		util.ChanSendEmpty(stopChan, 2)
		log.Debug("stop channel signaled, now waiting for port forward and sse streamer to exit")
	}

	wg.Wait()
	log.Debug("all goroutines have exited, now exiting data catcher")
	return nil
}

// consumeSSEStream connects to an SSE endpoint and processes events.
func (dc *DataCatcher) consumeSSEStream(ctx context.Context, url string, stopChan chan struct{}) error {
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
		case <-stopChan:
			log.Debug("stop signal received: in split func - sending back ErrFinalToken")
			return 0, nil, bufio.ErrFinalToken
		case <-ctx.Done():
			log.Debug("done signal received: in split func - sending back ErrFinalToken")
			return 0, nil, bufio.ErrFinalToken
		default:
			return bufio.ScanLines(data, atEOF)
		}
	})

	// Read the SSE stream line by line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse the event data
		if strings.HasPrefix(line, "data:") {
			log.Tracef("Stream data received: %s", line)
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			if err := dc.catcher(data); err != nil {
				return err
			}
		} else if strings.HasPrefix(line, "id:") || strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "retry:") {
			log.Debugf("SSE %s", line)
		} else if strings.HasPrefix(line, ":") {
			log.Debugf("SSE comment: %s", line)
		} else {
			log.Debugf("SSE unknown: %s", line)
			return ErrUnknownSSEData
		}
	}

	// Handle any errors during scanning
	err = scanner.Err()
	if err != nil {
		if util.IsErr(err, io.EOF, bufio.ErrFinalToken) {
			log.Debug("EOF or final token received from scanner in ConsumeSSEStream, exiting now")
		} else {
			return err
		}
	}
	return nil
}
