package logger

import (
	"github.com/inconshreveable/log15"
)

type log15LogEntry struct {
	entry log15.Logger
}
type log15Logger struct {
	logger log15.Logger
}

func newLog15Logger(config Configuration) (Logger, error) {
	return &log15Logger{
		logger: log15.New("", ""),
	}, nil
}

func (l *log15Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(format, args...)
}

func (l *log15Logger) Infof(format string, args ...interface{}) {
	l.logger.Info(format, args...)
}

func (l *log15Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn(format, args...)
}

func (l *log15Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error(format, args...)
}

func (l *log15Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Error(format, args...)
}

func (l *log15Logger) Panicf(format string, args ...interface{}) {
	l.logger.Error(format, args...)
}

func (l *log15Logger) WithFields(fields Fields) Logger {
	return &log15LogEntry{
		entry: l.logger.New(convertTolog15Fields(fields)),
	}
}

func (l *log15LogEntry) Debugf(format string, args ...interface{}) {
	l.entry.Debug(format, args...)
}

func (l *log15LogEntry) Infof(format string, args ...interface{}) {
	l.entry.Info(format, args...)
}

func (l *log15LogEntry) Warnf(format string, args ...interface{}) {
	l.entry.Warn(format, args...)
}

func (l *log15LogEntry) Errorf(format string, args ...interface{}) {
	l.entry.Error(format, args...)
}

func (l *log15LogEntry) Fatalf(format string, args ...interface{}) {
	l.entry.Error(format, args...)
}

func (l *log15LogEntry) Panicf(format string, args ...interface{}) {
	l.entry.Error(format, args...)
}

func (l *log15LogEntry) WithFields(fields Fields) Logger {
	return &log15LogEntry{
		entry: l.entry.New(convertTolog15Fields(fields)),
	}
}

func convertTolog15Fields(fields Fields) log15.Ctx {
	log15Fields := log15.Ctx{}
	for index, val := range fields {
		log15Fields[index] = val
	}
	return log15Fields
}
