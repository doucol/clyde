package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/doucol/clyde/internal/util"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var debugLog = os.Getenv("DEBUG_LOG") != ""

type LogMsg struct {
	ID      int    `json:"id" storm:"id,increment"`
	Message string `json:"message"`
}

type LogStore struct {
	db   *storm.DB
	msgs chan []byte
	wg   *sync.WaitGroup
}

func logPath() string {
	return filepath.Join(util.GetDataPath(), "logdata.db")
}

func New() (*LogStore, error) {
	lp := logPath()
	db, err := storm.Open(lp)
	if err != nil {
		return nil, err
	}
	err = db.Init(&LogMsg{})
	if err != nil {
		return nil, err
	}
	msgs := make(chan []byte, 1000)
	wg := &sync.WaitGroup{}
	ls := &LogStore{db: db, msgs: msgs, wg: wg}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for p := range msgs {
			if len(p) == 0 {
				return
			}
			if debugLog {
				_, err := os.Stderr.Write(p)
				runtime.HandleError(err)
			} else {
				lm := &LogMsg{Message: string(p)}
				runtime.HandleError(ls.db.Save(lm))
			}
		}
	}()
	return ls, nil
}

func Clear() error {
	lp := logPath()
	if util.FileExists(lp) {
		return os.Remove(lp)
	}
	return nil
}

func (l *LogStore) Stop() {
	l.msgs <- []byte{}
	l.wg.Wait()
}

func (l *LogStore) Close() error {
	return l.db.Close()
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

func (l *LogStore) Dump(w io.Writer) error {
	query := l.db.Select(q.Gte("ID", 1))
	err := query.Each(new(LogMsg), func(o any) error {
		lm := o.(*LogMsg)
		_, err := w.Write([]byte(lm.Message))
		return err
	})
	return err
}
