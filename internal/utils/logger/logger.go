// Пакет для логгирования
package logger

import (
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level int

// Level - Определяет уровни логгирования
const (
	DebugLevel Level = iota
	InfoLevel
	ErrorLevel
)

// Внешний интерфейс
type Logger interface {
	Info(...interface{})
	Debug(...interface{})
	Error(...interface{})
	LoggerEnabler
}

type LoggerEnabler interface {
	Enable()
	Disable()
}

type Log struct {
	Zap *zap.SugaredLogger
}

func ConvertLevel(level Level) zapcore.Level {
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

// NewLogger Создаёт логгер на основе zap.NewDevelopment()
func NewLogger(level Level) Logger {
	zapLevel := ConvertLevel(level)
	zap.NewAtomicLevelAt(zapLevel)
	logger, err := zap.NewDevelopment(zap.AddCallerSkip(2))
	if err != nil {
		return nil
	}
	log := Log{
		Zap: logger.Sugar(),
	}
	return &log
}

func (l Log) Info(args ...interface{}) {
	l.Zap.Infoln(args...)
}

func (l Log) Debug(args ...interface{}) {
	l.Zap.Debugln(args...)
}

func (l Log) Error(args ...interface{}) {
	if args[0] == nil {
		return
	}
	l.Zap.Errorln(args...)
}

// Disable Выключает логгирование
func (l *Log) Disable() {
	z := zap.NewNop()
	l.Zap = z.Sugar()
}

// Enable Включает логгирование
func (l *Log) Enable() {
	logger, err := zap.NewDevelopment(zap.AddCallerSkip(2))
	if err != nil {
		return
	}

	l.Zap = logger.Sugar()
}

func Disable() {
	StandardLogger.Disable()
}

func Enable() {
	StandardLogger.Enable()
}

// StandardLogger Стандартный логгер, чтоб не приходилось создавать новый
var StandardLogger Logger

func init() {
	StandardLogger = NewLogger(DebugLevel)
	StandardLogger.Disable()
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

func ZapInterceptor() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return zap.NewNop()
	}
	grpc_zap.ReplaceGrpcLogger(logger)
	return logger
}
