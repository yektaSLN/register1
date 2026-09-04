package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"time"

	"github.com/rs/zerolog"
)

var (
	file *os.File
	//for locking goroutines
	fileMu   sync.Mutex
	filePath string
)

// we build a logger so its output goes to stderr
func New(writer io.Writer) zerolog.Logger {
	return zerolog.New(writer).
		With().
		Timestamp().
		Logger()
}

func Configure() {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		panic(err)
	}

	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999999999Z07:00"
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(loc)
	}
}

// initializing the log file
func InitFile(path string) error {

	//until the file is still being changes no other goroutine is entered
	fileMu.Lock()

	//free mutex
	defer fileMu.Unlock()

	if file != nil {
		return nil
	}

	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	logFile, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}

	file = logFile
	filePath = path

	return nil
}

func WriteToFile(data []byte) error {
	fileMu.Lock()
	defer fileMu.Unlock()

	if file == nil {
		return nil
	}

	_, err := file.Write(data)

	return err
}

func FilePath() string {
	fileMu.Lock()
	defer fileMu.Unlock()

	return filePath
}

func CloseFile() error {
	fileMu.Lock()
	defer fileMu.Unlock()

	if file == nil {
		return nil
	}

	err := file.Close()

	file = nil

	return err
}
