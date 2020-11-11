package logger

import (
	reallog15 "github.com/inconshreveable/log15"
	"github.com/metal-stack/go-hal/internal/logger/log15"
	"github.com/metal-stack/go-hal/internal/logger/logrus"
	"github.com/metal-stack/go-hal/internal/logger/zap"
	reallogrus "github.com/sirupsen/logrus"
	uberzap "go.uber.org/zap"
)

// A global variable so that log functions can be directly accessed
var log Logger

// Logger is our contract for the logger
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
}

// New returns an simple instance of logger
func New() Logger {
	return log15.New(reallog15.New())
}

// NewZap returns an zap instance of logger
func NewZap(logger *uberzap.SugaredLogger) Logger {
	return zap.New(logger)
}

// NewLog15 returns an log15 instance of logger
func NewLog15(logger reallog15.Logger) Logger {
	return log15.New(logger)
}

// NewLogrus returns an logrus instance of logger
func NewLogrus(logger *reallogrus.Logger) Logger {
	return logrus.New(logger)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}
