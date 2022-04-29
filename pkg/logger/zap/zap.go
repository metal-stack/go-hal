package zap

import (
	"go.uber.org/zap"
)

type ZapLogger struct {
	sugaredLogger *zap.SugaredLogger
}

func New(logger *zap.SugaredLogger) *ZapLogger {
	return &ZapLogger{
		sugaredLogger: logger,
	}
}

func (l *ZapLogger) Debugw(format string, args ...any) {
	l.sugaredLogger.Debugw(format, args...)
}

func (l *ZapLogger) Infow(format string, args ...any) {
	l.sugaredLogger.Infow(format, args...)
}

func (l *ZapLogger) Warnw(format string, args ...any) {
	l.sugaredLogger.Warnw(format, args...)
}

func (l *ZapLogger) Errorw(format string, args ...any) {
	l.sugaredLogger.Errorw(format, args...)
}

func (l *ZapLogger) Fatalw(format string, args ...any) {
	l.sugaredLogger.Fatalw(format, args...)
}

func (l *ZapLogger) Panicw(format string, args ...any) {
	l.sugaredLogger.Fatalw(format, args...)
}
