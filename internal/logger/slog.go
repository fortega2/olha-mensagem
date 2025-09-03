package logger

import (
	"log/slog"
	"os"

	"github.com/fortega2/real-time-chat/internal/shutdown"
)

type slogLogger struct {
	l *slog.Logger
}

const (
	debugLevel string = "DEBUG"
	infoLevel  string = "INFO"
	warnLevel  string = "WARN"
	errorLevel string = "ERROR"
)

func NewSlogLogger() Logger {
	logLevel := getLevelInfo(os.Getenv("LOG_LEVEL"))

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	return &slogLogger{
		l: logger,
	}
}

func (s *slogLogger) Debug(msg string, args ...any) {
	s.l.Debug(msg, args...)
}

func (s *slogLogger) Info(msg string, args ...any) {
	s.l.Info(msg, args...)
}

func (s *slogLogger) Warn(msg string, args ...any) {
	s.l.Warn(msg, args...)
}

func (s *slogLogger) Error(msg string, args ...any) {
	s.l.Error(msg, args...)
}

func (s *slogLogger) Fatal(msg string, args ...any) {
	s.l.Error(msg, args...)
	shutdown.Signal()
}

func (s *slogLogger) With(args ...any) Logger {
	return &slogLogger{l: s.l.With(args...)}
}

func getLevelInfo(level string) slog.Level {
	switch level {
	case debugLevel:
		return slog.LevelDebug
	case infoLevel:
		return slog.LevelInfo
	case warnLevel:
		return slog.LevelWarn
	case errorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
