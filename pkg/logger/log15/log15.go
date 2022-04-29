package log15

import (
	"github.com/inconshreveable/log15"
)

type Log15Logger struct {
	logger log15.Logger
}

func New(logger log15.Logger) *Log15Logger {
	return &Log15Logger{
		logger: logger,
	}
}

func (l *Log15Logger) Debugw(format string, args ...any) {
	l.logger.Debug(format, args...)
}

func (l *Log15Logger) Infow(format string, args ...any) {
	l.logger.Info(format, args...)
}

func (l *Log15Logger) Warnw(format string, args ...any) {
	l.logger.Warn(format, args...)
}

func (l *Log15Logger) Errorw(format string, args ...any) {
	l.logger.Error(format, args...)
}

func (l *Log15Logger) Fatalw(format string, args ...any) {
	l.logger.Error(format, args...)
}

func (l *Log15Logger) Panicw(format string, args ...any) {
	l.logger.Error(format, args...)
}
