package logger

import (
	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	ErrorLevel
)

type Logger interface {
	Info(...interface{})
	Debug(...interface{})
	Error(...interface{})
}

type Log struct {
	Zap *zap.SugaredLogger
}

func chooseLevel(level Level) zapcore.Level {
	var zapLevel zapcore.Level
	switch level {
	case DebugLevel:
		zapLevel = zap.DebugLevel
	case InfoLevel:
		zapLevel = zap.InfoLevel
	case ErrorLevel:
		zapLevel = zap.ErrorLevel
	}
	return zapLevel
}

func NewLogger(level Level) Logger {
	zapLevel := chooseLevel(level)
	zap.NewAtomicLevelAt(zapLevel)
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	log := Log{
		Zap: logger.Sugar(),
	}
	return log
}

func (l Log) Info(args ...interface{}) {
	color.Blue("[INFO] ")
	l.Zap.Infoln(args...)
}

func (l Log) Debug(args ...interface{}) {
	color.Green("[DEBUG] ")
	l.Zap.Debugln(args...)
}

func (l Log) Error(args ...interface{}) {
	color.Red("[ERROR] ")
	l.Zap.Errorln(args...)
}

var StandardLogger Logger

func init() {
	StandardLogger = NewLogger(DebugLevel)
}

func Info(args ...interface{}) {
	StandardLogger.Info(args...)
}

func Debug(args ...interface{}) {
	StandardLogger.Debug(args...)
}

func Error(args ...interface{}) {
	StandardLogger.Error(args...)
}
