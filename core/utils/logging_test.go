package utils

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestConfigureLoggerUsesGlobalDefault(t *testing.T) {
	var output bytes.Buffer

	logger := ConfigureLogger(LoggerConfig{
		Level:  slog.LevelDebug,
		Format: LogFormatJSON,
		Writer: &output,
	})

	if Logger() != logger {
		t.Fatal("expected configured logger to become the global default")
	}

	Debug("configured logger works", "component", "test")

	logLine := output.String()
	if !strings.Contains(logLine, `"level":"DEBUG"`) {
		t.Fatalf("expected debug level in log output, got %q", logLine)
	}
	if !strings.Contains(logLine, `"component":"test"`) {
		t.Fatalf("expected structured field in log output, got %q", logLine)
	}
}

func TestParseLogLevel(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    slog.Level
		wantErr bool
	}{
		{name: "empty defaults to info", input: "", want: slog.LevelInfo},
		{name: "debug", input: "debug", want: slog.LevelDebug},
		{name: "info", input: "info", want: slog.LevelInfo},
		{name: "warn", input: "warn", want: slog.LevelWarn},
		{name: "warning alias", input: "warning", want: slog.LevelWarn},
		{name: "error", input: "error", want: slog.LevelError},
		{name: "invalid", input: "trace", wantErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := ParseLogLevel(testCase.input)
			if testCase.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != testCase.want {
				t.Fatalf("expected %v, got %v", testCase.want, got)
			}
		})
	}
}
