// the logger module provides structured log data output using ZAP
package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLog *zap.Logger

func init() {
	var err error
	//config := zap.NewProductionConfig()
	//encoderConfig := zap.NewProductionEncoderConfig()
	config := zap.NewDevelopmentConfig()
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05")
	encoderConfig.CallerKey = ""     // to hide caller line key info
	encoderConfig.StacktraceKey = "" // to hide stacktrace info
	config.EncoderConfig = encoderConfig

	zapLog, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		//"srv: cannot initialize zap logger"
		panic(err)
	}
}

// provides info level logging
func Info(message string, fields ...zap.Field) {
	zapLog.Info(message, fields...)
}

// provides debug level logging
func Debug(message string, fields ...zap.Field) {
	zapLog.Debug(message, fields...)
}

// provides error level logging
func Error(message string, fields ...zap.Field) {
	zapLog.Error(message, fields...)
}

// provides fatal level logging
func Fatal(message string, fields ...zap.Field) {
	zapLog.Fatal(message, fields...)
}

// provides info level logging with Sprintf functionality
func Infof(format string, a ...any) {
	zapLog.Info(fmt.Sprintf(format, a...))
}

// provides debug level logging with Sprintf functionality
func Debugf(format string, a ...any) {
	zapLog.Debug(fmt.Sprintf(format, a...))
}

// provides error level logging with Sprintf functionality
func Errorf(format string, a ...any) {
	zapLog.Error(fmt.Sprintf(format, a...))
}

// provides fatal level logging with Sprintf functionality
func Fatalf(format string, a ...any) {
	zapLog.Fatal(fmt.Sprintf(format, a...))
}

// standardized record function for command trace logging
func CommandTrace(method string, path string, status int, duration time.Duration) {
	zapLog.Info("Incoming request",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", status),
		zap.Duration("duration", duration),
	)
}

// does log sync to prevent buffered data loss
func Sync() {
	zapLog.Sync()
}
