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

func (l *Log15Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(format, args...)
}

func (l *Log15Logger) Infof(format string, args ...interface{}) {
	l.logger.Info(format, args...)
}

func (l *Log15Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn(format, args...)
}

func (l *Log15Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error(format, args...)
}

func (l *Log15Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Error(format, args...)
}

func (l *Log15Logger) Panicf(format string, args ...interface{}) {
	l.logger.Error(format, args...)
}
