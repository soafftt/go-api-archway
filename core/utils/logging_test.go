package utils

import (
	"testing"
)

func TestGetLogger_ReturnsNonNilLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Fatal("expected non-nil logger from GetLogger")
	}
}

func TestLogger_Methods_DoNotPanic(t *testing.T) {
	logger := GetLogger()

	logger.Debug("debug message")
	logger.DebugW("debugw message", "key", "value")
	logger.Info("info message")
	logger.InfoW("infow message", "key", "value")
}
