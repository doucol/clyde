package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/doucol/clyde/internal/util"
)

type LogMsg struct {
	ID      int    `json:"id" storm:"id,increment"`
	Message string `json:"message"`
}

type LogStore struct {
	db *storm.DB
	mu sync.Mutex
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
	return &LogStore{db: db}, nil
}

func Clear() error {
	lp := logPath()
	if util.FileExists(lp) {
		return os.Remove(lp)
	}
	return nil
}

func (l *LogStore) Close() error {
	return l.db.Close()
}

// [Writer] interface
func (l *LogStore) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	err = l.db.Save(&LogMsg{Message: string(p)})
	if err != nil {
		return 0, err
	}
	return len(p), nil
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
