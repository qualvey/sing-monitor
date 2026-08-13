package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "info", "":
		return LevelInfo
	default:
		return LevelInfo
	}
}

var (
	defaultWriter io.Writer = os.Stdout
	std                     = log.New(defaultWriter, "", log.LstdFlags)
	mu            sync.RWMutex
	currentLevel  = LevelInfo
)

func Init(level string) {
	mu.Lock()
	defer mu.Unlock()
	currentLevel = ParseLevel(level)
	std.SetPrefix("")
	std.SetFlags(log.LstdFlags)
}

func enabled(level Level) bool {
	mu.RLock()
	defer mu.RUnlock()
	return level >= currentLevel
}

func logf(level Level, format string, args ...interface{}) {
	if !enabled(level) {
		return
	}
	std.Printf("[%s] %s", level.String(), fmt.Sprintf(format, args...))
}

func Debug(format string, args ...interface{}) {
	logf(LevelDebug, format, args...)
}

func Info(format string, args ...interface{}) {
	logf(LevelInfo, format, args...)
}

func Warn(format string, args ...interface{}) {
	logf(LevelWarn, format, args...)
}

func Error(format string, args ...interface{}) {
	logf(LevelError, format, args...)
}

func Fatal(format string, args ...interface{}) {
	logf(LevelError, format, args...)
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	Fatal(format, args...)
}

func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	std.SetOutput(w)
}

func ResetOutput() {
	mu.Lock()
	defer mu.Unlock()
	std.SetOutput(defaultWriter)
}
