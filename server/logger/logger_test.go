package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]Level{
		"debug":   LevelDebug,
		"info":    LevelInfo,
		"warn":    LevelWarn,
		"error":   LevelError,
		"warning": LevelWarn,
		"INFO":    LevelInfo,
	}

	for in, want := range cases {
		if got := ParseLevel(in); got != want {
			t.Fatalf("ParseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestInitHonorsLevel(t *testing.T) {
	oldLevel := currentLevel
	buf := &bytes.Buffer{}
	std.SetOutput(buf)
	defer std.SetOutput(defaultWriter)
	defer func() { currentLevel = oldLevel }()

	Init("warn")
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	out := buf.String()
	if strings.Contains(out, "debug message") {
		t.Fatalf("debug message should be filtered out at warn level: %s", out)
	}
	if strings.Contains(out, "info message") {
		t.Fatalf("info message should be filtered out at warn level: %s", out)
	}
	if !strings.Contains(out, "warn message") || !strings.Contains(out, "error message") {
		t.Fatalf("warn and error messages should remain: %s", out)
	}
}
