package logger

import (
	"go.uber.org/zap"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var Log *zap.SugaredLogger

func init() {
	log := zap.NewExample()
	defer log.Sync()
	Log = log.Sugar()
}

func Info(format string, args ...any) {
	Log.Infof(format+"\n", args...)
}

func Debug(format string, args ...any) {
	Log.Debugf(format+"\n", args...)
}

func Error(format string, args ...any) {
	Log.Errorf(format+"\n", args...)
}
