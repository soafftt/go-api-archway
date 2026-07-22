package utils

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var _logger Logger

// 로거를 읽는 방법은 이 방법이 최선이 아닐까 싶음. 라고 생각하지만 아닐 수도 있음
func init() {
	_ = godotenv.Load()
	cfg := &config{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	_logger = newLogger(cfg.Environment)
}

type Logger interface {
	Debug(message string)
	DebugW(message string, keyValue ...any)
	Info(message string)
	InfoW(message string, keyValue ...any)
	Warn(message string)
	WarnW(message string, err error, keyValue ...any)
	Error(message string)
	ErrorW(message string, err error, keyValue ...any)
	Sync()
}

func GetLogger() Logger {
	return _logger
}

type config struct {
	Environment string `env:"ENV" envDefault:"dev"`
}

func newLogger(environment string) Logger {
	var logger *zap.Logger
	if environment == "local" || environment == "dev" {
		logger, _ = zap.NewDevelopment()

	} else {
		logger, _ = zap.NewProduction()
	}

	logger = logger.WithOptions(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.WarnLevel),
	)

	return &stdLogger{
		zapLogger: logger,
		zapSugaredLogger: logger.Sugar().WithOptions(
			zap.AddCaller(),
			zap.AddCallerSkip(1),
			zap.AddStacktrace(zap.WarnLevel),
		),
	}
}

type stdLogger struct {
	zapLogger        *zap.Logger
	zapSugaredLogger *zap.SugaredLogger
}

func (s stdLogger) Debug(message string) {
	s.zapLogger.Debug(message)
}

func (s stdLogger) DebugW(message string, keyValue ...any) {
	s.zapSugaredLogger.Debugw(message, keyValue...)
}

func (s stdLogger) Info(message string) {
	s.zapSugaredLogger.Info(message)
}

func (s stdLogger) InfoW(message string, keyValue ...any) {
	s.zapSugaredLogger.Infow(message, keyValue...)
}

func (s stdLogger) Warn(message string) {
	s.zapSugaredLogger.Warn(message)
}

func (s stdLogger) WarnW(message string, err error, keyValue ...any) {
	s.zapSugaredLogger.Warnw(message, append(keyValue, "error", err)...)
}

func (s stdLogger) Error(message string) {
	s.zapSugaredLogger.Error(message)
}

func (s stdLogger) ErrorW(message string, err error, keyValue ...any) {
	s.zapSugaredLogger.Errorw(message, append(keyValue, "error", err)...)
}

func (s stdLogger) Sync() {
	_ = s.zapSugaredLogger.Sync()
	_ = s.zapLogger.Sync()
}
