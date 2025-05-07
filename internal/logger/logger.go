package logger

import (
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
	msgs := make(chan []byte, 1000)
	wg := &sync.WaitGroup{}
	ls := &Logger{
		wg:   wg,
		msgs: msgs,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY | os.O_APPEND
		lf, err := os.OpenFile(lp, flags, 0666)
		if err != nil {
			panic(err)
		}
		defer lf.Close()
		for p := range msgs {
			if len(p) == 0 {
				return
			}
			_, err := lf.Write(p)
			if err != nil {
				panic(err)
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

func (l *Logger) Dump(to io.Writer) {
	lf, err := os.Open(GetLogFile())
	if err != nil {
		panic(err)
	}
	defer lf.Close()
	_, err = lf.WriteTo(to)
	if err != nil {
		panic(err)
	}
}
