package logger

import (
	"log/slog"
	"os"

	halslog "github.com/metal-stack/go-hal/pkg/logger/slog"
)

// A global variable so that log functions can be directly accessed
var log Logger

// Logger is our contract for the logger
type Logger interface {
	Debugw(format string, args ...any)
	Infow(format string, args ...any)
	Warnw(format string, args ...any)
	Errorw(format string, args ...any)
	Fatalw(format string, args ...any)
	Panicw(format string, args ...any)
}

// New returns an simple instance of logger
func New() Logger {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	log := slog.New(jsonHandler)
	return halslog.New(log)
}

// NewSlog returns an zap instance of logger
func NewSlog(logger *slog.Logger) Logger {
	return halslog.New(logger)
}

func Debugw(format string, args ...any) {
	log.Debugw(format, args...)
}

func Infow(format string, args ...any) {
	log.Infow(format, args...)
}

func Warnw(format string, args ...any) {
	log.Warnw(format, args...)
}

func Errorw(format string, args ...any) {
	log.Errorw(format, args...)
}

func Fatalw(format string, args ...any) {
	log.Fatalw(format, args...)
}

func Panicw(format string, args ...any) {
	log.Panicw(format, args...)
}
