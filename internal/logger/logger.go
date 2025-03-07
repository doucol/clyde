package logger

import (
	"errors"
	"io"
	"os"
	"path/filepath"

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
}

func logPath() string {
	return filepath.Join(util.GetDataPath(), "logdata.db")
}

func New() (*LogStore, error) {
	fp := logPath()
	db, err := storm.Open(fp)
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
	fp := logPath()
	if _, err := os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return os.Remove(fp)
}

func (l *LogStore) Close() error {
	return l.db.Close()
}

func (l *LogStore) Write(p []byte) (n int, err error) {
	err = l.db.Save(&LogMsg{Message: string(p)})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (l *LogStore) Dump(w io.Writer) error {
	if l.db == nil {
		return nil
	}
	query := l.db.Select(q.Gte("ID", 1))
	err := query.Each(new(LogMsg), func(o any) error {
		lm := o.(*LogMsg)
		_, err := w.Write([]byte(lm.Message))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
