package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/doucol/clyde/internal/util"
)

func TestLoggerCreation(t *testing.T) {
	// Test with default log file
	logger, err := NewLogger()
	if err != nil {
		t.Errorf("NewLogger() error = %v", err)
	}
	if logger == nil {
		t.Error("NewLogger() returned nil logger")
	}
	logger.Close()

	// Test with custom log file
	testLogFile := filepath.Join(os.TempDir(), "test.log")
	SetLogFile(testLogFile)
	defer os.Remove(testLogFile)

	logger, err = NewLogger()
	if err != nil {
		t.Errorf("NewLogger() with custom file error = %v", err)
	}
	if logger == nil {
		t.Error("NewLogger() with custom file returned nil logger")
	}
	logger.Close()
}

func TestLoggerWrite(t *testing.T) {
	// Create a temporary log file
	testLogFile := filepath.Join(os.TempDir(), "test.log")
	SetLogFile(testLogFile)
	defer os.Remove(testLogFile)

	logger, err := NewLogger()
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Close()

	// Test writing messages
	testMessages := []string{
		"Test message 1\n",
		"Test message 2\n",
		"Test message 3\n",
	}

	for _, msg := range testMessages {
		n, err := logger.Write([]byte(msg))
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
		if n != len(msg) {
			t.Errorf("Write() wrote %d bytes; want %d", n, len(msg))
		}
	}

	// Give some time for the messages to be written
	time.Sleep(100 * time.Millisecond)

	// Verify the contents
	content, err := os.ReadFile(testLogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	expectedContent := ""
	for _, msg := range testMessages {
		expectedContent += msg
	}

	if string(content) != expectedContent {
		t.Errorf("Log file content = %q; want %q", string(content), expectedContent)
	}
}

func TestLoggerDump(t *testing.T) {
	// Create a temporary log file
	testLogFile := filepath.Join(os.TempDir(), "test.log")
	SetLogFile(testLogFile)
	defer os.Remove(testLogFile)

	logger, err := NewLogger()
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Close()

	// Write some test messages
	testMessages := []string{
		"Test message 1\n",
		"Test message 2\n",
		"Test message 3\n",
	}

	for _, msg := range testMessages {
		logger.Write([]byte(msg))
	}

	// Give some time for the messages to be written
	time.Sleep(100 * time.Millisecond)

	// Test dumping to a buffer
	var buf bytes.Buffer
	logger.Dump(&buf)

	expectedContent := ""
	for _, msg := range testMessages {
		expectedContent += msg
	}

	if buf.String() != expectedContent {
		t.Errorf("Dump() content = %q; want %q", buf.String(), expectedContent)
	}
}

func TestLogFilePaths(t *testing.T) {
	// Test default log file path
	defaultPath := GetDefaultLogFile()
	expectedDefaultPath := filepath.Join(util.GetDataPath(), "clyde.log")
	if defaultPath != expectedDefaultPath {
		t.Errorf("GetDefaultLogFile() = %v; want %v", defaultPath, expectedDefaultPath)
	}

	// Test custom log file path
	customPath := "/custom/path/to/log.log"
	SetLogFile(customPath)
	if GetLogFile() != customPath {
		t.Errorf("GetLogFile() = %v; want %v", GetLogFile(), customPath)
	}

	// Test getting default path when no custom path is set
	SetLogFile("")
	if GetLogFile() != expectedDefaultPath {
		t.Errorf("GetLogFile() with empty path = %v; want %v", GetLogFile(), expectedDefaultPath)
	}
}

func TestLoggerConcurrentWrite(t *testing.T) {
	// Create a temporary log file
	testLogFile := filepath.Join(os.TempDir(), "test.log")
	SetLogFile(testLogFile)
	defer os.Remove(testLogFile)

	logger, err := NewLogger()
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Close()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			msg := []byte("Concurrent message " + string(rune(id+'0')) + "\n")
			logger.Write(msg)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Give some time for the messages to be written
	time.Sleep(100 * time.Millisecond)

	// Verify the contents
	content, err := os.ReadFile(testLogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Check if all messages are present
	for i := 0; i < 10; i++ {
		expectedMsg := "Concurrent message " + string(rune(i+'0')) + "\n"
		if !bytes.Contains(content, []byte(expectedMsg)) {
			t.Errorf("Log file missing message: %q", expectedMsg)
		}
	}
}
