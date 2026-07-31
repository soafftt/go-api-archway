package utils

import (
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _logger Logger

// 로거를 읽는 방법은 이 방법이 최선이 아닐까 싶음. 라고 생각하지만 아닐 수도 있음
func init() {
	_ = godotenv.Load()
	cfg := &config{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	_logger = newLogger(cfg.Environment, cfg.Logger.LogLevel)
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
	Logger      struct {
		LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	}
}

func newLogger(environment string, logLevel string) Logger {
	var logEncoderConfig zapcore.EncoderConfig
	if environment == "local" || environment == "dev" {
		logEncoderConfig = zap.NewDevelopmentEncoderConfig()

	} else {
		logEncoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	// 레벨표시를 Color 사용.
	logEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(logEncoderConfig), // 콘솔용 인코더
		zapcore.AddSync(os.Stdout),                  // 표준 출력
		getLogLevel(logLevel),                       // 로그 레벨
	)

	logger := zap.New(
		core,
		// Wrapping 한 상태이기에 Wrapping 위치를 제거하기 위하여 1단계 스킵
		zap.AddCallerSkip(1),
		// Warn 부터는 Stacktrace 추가.
		zap.AddStacktrace(zap.WarnLevel),
	)

	return &stdLogger{
		zapLogger: logger,
		zapSugaredLogger: logger.Sugar().WithOptions(
			zap.AddCaller(),
			zap.AddCallerSkip(1),
			zap.AddStacktrace(zap.WarnLevel),
			zap.ErrorOutput(os.Stdout),
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

func getLogLevel(lv string) zapcore.Level {
	var logLv zapcore.Level
	switch strings.ToLower(lv) {
	case "INFO":
		logLv = zap.InfoLevel
	case "WRAN":
		logLv = zap.WarnLevel
	case "ERROR":
		logLv = zap.ErrorLevel
	case "DPNIC":
	case "PANIC":
		logLv = zap.PanicLevel

	case "FATAL":
		logLv = zap.FatalLevel
	default:
		logLv = zap.DebugLevel
	}

	return logLv
}
