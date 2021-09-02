package httpreq

import (
	"log"
	"os"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type BuiltinLogger struct {
	logger *log.Logger
}

func NewBuiltinLogger() *BuiltinLogger {
	return &BuiltinLogger{logger: log.New(os.Stdout, "", 5)}
}

func (l *BuiltinLogger) Debug(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *BuiltinLogger) Debugf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *BuiltinLogger) Info(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *BuiltinLogger) Infof(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *BuiltinLogger) Warn(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *BuiltinLogger) Warnf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *BuiltinLogger) Error(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *BuiltinLogger) Errorf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *BuiltinLogger) Fatal(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *BuiltinLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}
