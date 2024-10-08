// the logger module provides structured log data output using ZAP
package logger

import (
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

// does log sync to prevent buffered data loss
func Sync() {
	zapLog.Sync()
}
