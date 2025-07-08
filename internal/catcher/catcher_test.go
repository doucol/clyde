package catcher

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockCatcher is a simple CatcherFunc for testing
func mockCatcher(collected *[]string) CatcherFunc {
	return func(data string) error {
		*collected = append(*collected, data)
		return nil
	}
}

func TestDataCatcher_CatchServerSentEvents_WithURLFull(t *testing.T) {
	var received []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: testdata\n\n"))
	}))
	defer server.Close()

	dc := &DataCatcher{
		catcher:     mockCatcher(&received),
		urlPath:     "/events",
		URLFull:     server.URL,
		recoverFunc: func() {},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sseReady := make(chan bool, 1)
	err := dc.CatchServerSentEvents(ctx, sseReady)
	if err != nil {
		t.Fatalf("CatchServerSentEvents returned error: %v", err)
	}
	if len(received) == 0 || received[0] != "testdata" {
		t.Errorf("Expected to receive 'testdata', got: %v", received)
	}
}

func TestDataCatcher_consumeSSEStream_HandlesNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer server.Close()
	dc := &DataCatcher{
		catcher:     func(data string) error { return nil },
		recoverFunc: func() {},
	}
	err := dc.consumeSSEStream(context.Background(), server.URL, make(chan struct{}), make(chan bool, 1))
	if err == nil || !errors.Is(err, io.EOF) && err.Error() != "unexpected status code: 418" {
		t.Errorf("Expected error for non-200 status, got: %v", err)
	}
}

func TestDataCatcher_consumeSSEStream_HandlesDataAndOtherLines(t *testing.T) {
	var received []string
	body := "id: 1\nevent: test\ndata: foo\nmessage: bar\ndata: bar\n\n"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
	defer ts.Close()
	dc := &DataCatcher{
		catcher:     mockCatcher(&received),
		recoverFunc: func() {},
	}
	err := dc.consumeSSEStream(context.Background(), ts.URL, make(chan struct{}), make(chan bool, 1))
	if err != nil {
		t.Fatalf("consumeSSEStream returned error: %v", err)
	}
	if len(received) != 2 || received[0] != "foo" || received[1] != "bar" {
		t.Errorf("Expected to receive ['foo', 'bar'], got: %v", received)
	}
}

func TestDataCatcher_consumeSSEStream_HandlesCatcherError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: fail\n\n"))
	}))
	defer ts.Close()
	dc := &DataCatcher{
		catcher: func(data string) error {
			return errors.New("catcher error")
		},
		recoverFunc: func() {},
	}
	err := dc.consumeSSEStream(context.Background(), ts.URL, make(chan struct{}), make(chan bool, 1))
	if err == nil || err.Error() != "catcher error" {
		t.Errorf("Expected catcher error, got: %v", err)
	}
}

func TestDataCatcher_CatchServerSentEvents_ContextCancel(t *testing.T) {
	dc := &DataCatcher{
		catcher:     func(data string) error { return nil },
		recoverFunc: func() {},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sseReady := make(chan bool, 1)
	err := dc.CatchServerSentEvents(ctx, sseReady)
	if err != nil {
		t.Errorf("Expected nil error when context is cancelled, got: %v", err)
	}
}

func TestDataCatcher_consumeSSEStream_StopChan(t *testing.T) {
	var received []string
	body := "data: foo\n\n"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
	defer ts.Close()
	dc := &DataCatcher{
		catcher:     mockCatcher(&received),
		recoverFunc: func() {},
	}
	stopChan := make(chan struct{}, 1)
	stopChan <- struct{}{}
	err := dc.consumeSSEStream(context.Background(), ts.URL, stopChan, make(chan bool, 1))
	if err != nil {
		t.Errorf("Expected nil error when stopChan is closed, got: %v", err)
	}
}
