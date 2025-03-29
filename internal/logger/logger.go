package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/doucol/clyde/internal/util"
)

type LogStore struct {
	msgs chan []byte
	wg   *sync.WaitGroup
}

func logPath() string {
	return filepath.Join(util.GetDataPath(), "clyde.log")
}

func New() (*LogStore, error) {
	lp := logPath()
	msgs := make(chan []byte, 1000)
	wg := &sync.WaitGroup{}
	ls := &LogStore{
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

func (l *LogStore) Close() {
	l.msgs <- []byte{}
	l.wg.Wait()
}

// [Writer] interface
func (l *LogStore) Write(p []byte) (n int, err error) {
	length := len(p)
	if length > 0 {
		msgb := make([]byte, length)
		copy(msgb, p)
		l.msgs <- msgb
	}
	return length, nil
}

func (l *LogStore) Dump(to io.Writer) {
	lf, err := os.Open(logPath())
	if err != nil {
		panic(err)
	}
	defer lf.Close()
	_, err = lf.WriteTo(to)
	if err != nil {
		panic(err)
	}
}
