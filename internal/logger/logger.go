package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/doucol/clyde/internal/util"
)

type Logger struct {
	msgs chan []byte
	wg   *sync.WaitGroup
}

var logFile string

func SetLogFile(lf string) {
	logFile = lf
}

func GetLogFile() string {
	if logFile == "" {
		return GetDefaultLogFile()
	}
	return logFile
}

func GetDefaultLogFile() string {
	return filepath.Join(util.GetDataPath(), "clyde.log")
}

func NewLogger() (*Logger, error) {
	lp := GetLogFile()
	// Open the file synchronously so a failure (e.g. an unwritable path passed
	// via --logfile) is returned to the caller instead of panicking in a
	// goroutine that would crash the whole process.
	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	lf, err := os.OpenFile(lp, flags, 0o644)
	if err != nil {
		return nil, err
	}
	msgs := make(chan []byte, 1000)
	wg := &sync.WaitGroup{}
	ls := &Logger{
		wg:   wg,
		msgs: msgs,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if cerr := lf.Close(); cerr != nil {
				fmt.Fprintf(os.Stderr, "clyde: error closing log file: %v\n", cerr)
			}
		}()
		for p := range msgs {
			if len(p) == 0 {
				return
			}
			if _, err := lf.Write(p); err != nil {
				fmt.Fprintf(os.Stderr, "clyde: error writing to log file: %v\n", err)
				return
			}
		}
	}()
	return ls, nil
}

func (l *Logger) Close() {
	l.msgs <- []byte{}
	l.wg.Wait()
}

// [io.Writer] interface
func (l *Logger) Write(p []byte) (n int, err error) {
	length := len(p)
	if length > 0 {
		msgb := make([]byte, length)
		copy(msgb, p)
		l.msgs <- msgb
	}
	return length, nil
}

func (l *Logger) Dump(to io.Writer) error {
	lf, err := os.Open(GetLogFile())
	if err != nil {
		return err
	}
	defer func() { _ = lf.Close() }()
	_, err = lf.WriteTo(to)
	return err
}
