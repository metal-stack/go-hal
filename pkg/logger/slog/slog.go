package slog

import (
	"fmt"
	"log/slog"
)

type SlogLogger struct {
	log *slog.Logger
}

func New(logger *slog.Logger) *SlogLogger {
	return &SlogLogger{
		log: logger,
	}
}

func (l *SlogLogger) Debugw(format string, args ...any) {
	l.log.Debug(format, args...)
}

func (l *SlogLogger) Infow(format string, args ...any) {
	l.log.Info(format, args...)
}

func (l *SlogLogger) Warnw(format string, args ...any) {
	l.log.Warn(format, args...)
}

func (l *SlogLogger) Errorw(format string, args ...any) {
	l.log.Error(format, args...)
}

func (l *SlogLogger) Fatalw(format string, args ...any) {
	l.log.Error(format, args...)
	panic(fmt.Errorf(format, args...))
}

func (l *SlogLogger) Panicw(format string, args ...any) {
	l.log.Error(format, args...)
	panic(fmt.Errorf(format, args...))
}
