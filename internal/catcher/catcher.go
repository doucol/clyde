// Package catcher provides functionality to catch data from a server-sent events (SSE) stream
package catcher

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/util"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type pfOut struct{}

func (l pfOut) Write(bytes []byte) (int, error) {
	logrus.Debugf("portforward stdout: %s", string(bytes))
	return len(bytes), nil
}

type pfErr struct{}

func (l pfErr) Write(bytes []byte) (int, error) {
	logrus.Debugf("portforward stderr: %s", string(bytes))
	return len(bytes), nil
}

type CatcherFunc func(data string) error

type DataCatcher struct {
	catcher        CatcherFunc
	namespace      string
	containerName  string
	urlPath        string
	PortEnvVarName string
	recoverFunc    func()
	URLFull        string
}

func NewDataCatcher(namespace, containerName, urlPath string, catcher CatcherFunc, recover func()) *DataCatcher {
	return &DataCatcher{
		namespace:      namespace,
		containerName:  containerName,
		urlPath:        urlPath,
		catcher:        catcher,
		PortEnvVarName: "PORT",
		recoverFunc:    recover,
	}
}

func (dc *DataCatcher) CatchServerSentEvents(ctx context.Context, sseReady chan bool) error {
	select {
	case <-ctx.Done():
		logrus.Debug("done signal received - not entering CatchDataFromSSEStream")
		return nil
	default:
		logrus.Debug("entering data catcher")
	}

	wg := &sync.WaitGroup{}

	// Channels for port forward signaling
	stopChan := make(chan struct{}, 20)
	readyChan := make(chan struct{}, 5)

	defer func() {
		util.ChanClose(stopChan, readyChan)
		logrus.Debug("exited data catcher")
	}()

	var err error
	sseURL := dc.URLFull
	if sseURL == "" {
		sseURL, err = dc.portFoward(ctx, stopChan, readyChan, wg)
		if err != nil {
			return err
		}
	} else {
		readyChan <- struct{}{}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer util.ChanSendEmpty(stopChan, 2)
		defer dc.recoverFunc()
		select {
		case <-readyChan:
			// Wait for the port forwarding to be ready
			logrus.Debug("SSE server is ready, now starting SSE consumer")
			if err := dc.consumeSSEStream(ctx, sseURL, stopChan, sseReady); err != nil {
				logrus.Debugf("error: ConsumeSSEStream return error: %s", err.Error())
			}
		case <-time.After(5 * time.Second):
			logrus.Debug("timeout waiting for port forward to be ready")
		}
		logrus.Debug("sse consumer has stopped")
	}()

	select {
	case <-ctx.Done():
		util.ChanSendEmpty(stopChan, 2)
		logrus.Debug("done signal received, now waiting for port forward and sse consumer to exit")
	case <-stopChan:
		util.ChanSendEmpty(stopChan, 2)
		logrus.Debug("stop channel signaled, now waiting for port forward and sse consumer to exit")
	}

	wg.Wait()
	logrus.Debug("all goroutines have exited, now exiting data catcher")
	return nil
}

func (dc *DataCatcher) portFoward(ctx context.Context, stopChan, readyChan chan struct{}, wg *sync.WaitGroup) (string, error) {
	config := cmdctx.K8sConfigFromContext(ctx)
	clientset := cmdctx.K8sClientsetFromContext(ctx)

	// URL for the portforward endpoint on the pod
	podName, port, err := util.GetPodAndEnvVarByContainerName(ctx, clientset, dc.namespace, dc.containerName, dc.PortEnvVarName)
	if err != nil {
		return "", err
	}

	apiURL, _ := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", config.Host, dc.namespace, podName))

	// Dialer for establishing the connection
	logrus.Debugf("apiURL: %s", apiURL)
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return "", err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiURL)

	// Port mappings (local port: pod port)
	freePort, err := util.GetFreePort()
	if err != nil {
		return "", err
	}
	ports := []string{fmt.Sprintf("%d:%s", freePort, port)}

	pf, err := portforward.New(dialer, ports, stopChan, readyChan, pfOut{}, pfErr{})
	if err != nil {
		return "", err
	}

	// Start the port forwarding
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer util.ChanSendEmpty(stopChan, 2)
		defer dc.recoverFunc()
		logrus.Debugf("Starting port forward from localhost:%d to %s/%s:%s", freePort, dc.namespace, podName, port)
		if err := pf.ForwardPorts(); err != nil {
			logrus.Debugf("error: ForwardPorts return error: %s", err.Error())
		}
		logrus.Debug("port forward has stopped")
	}()

	sseURL := fmt.Sprintf("http://localhost:%d%s", freePort, dc.urlPath)
	return sseURL, nil
}

// consumeSSEStream connects to an SSE endpoint and processes events.
func (dc *DataCatcher) consumeSSEStream(ctx context.Context, url string, stopChan chan struct{}, sseReady chan bool) error {
	logrus.Debugf("Connecting to SSE stream at %s", url)
	util.ChanClose(sseReady)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to SSE stream: %w", err)
	}
	logrus.Debug("SSE stream responded")
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var scanDone chan bool
	defer util.ChanClose(scanDone)
	scanner := bufio.NewScanner(resp.Body)
	ssePrefixes := []string{"id:", "event:", "retry:", "message:", ":"}

	// Read the SSE stream line by line
	logrus.Debug("Entering SSE stream consumer loop")
	for {
		util.ChanClose(scanDone)
		scanDone = make(chan bool, 1)
		go func() {
			defer dc.recoverFunc()
			scanDone <- scanner.Scan()
		}()

		select {
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return nil
		case hasMore := <-scanDone:
			if !hasMore {
				return nil
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			// Parse the event data
			// We don't handle multi-line events yet i.e. 'data: foo'\n'data: continuation of foo`
			if strings.HasPrefix(line, "data:") {
				logrus.Tracef("Stream data received: %s", line)
				data := strings.TrimPrefix(line, "data:")
				data = strings.TrimSpace(data)
				if err := dc.catcher(data); err != nil {
					return err
				}
			} else {
				prefix := "unknown:"
				for _, p := range ssePrefixes {
					if strings.HasPrefix(line, p) {
						prefix = p
						break
					}
				}
				logrus.Debugf("SSE %s %s", prefix, line)
			}
			// Handle any errors during scanning
			err = scanner.Err()
			if err != nil {
				if util.IsErr(err, io.EOF, bufio.ErrFinalToken, io.ErrUnexpectedEOF) {
					logrus.Debug("EOF or final token received from scanner in ConsumeSSEStream, exiting now")
					return nil
				} else {
					return err
				}
			}
		}
	}
}
