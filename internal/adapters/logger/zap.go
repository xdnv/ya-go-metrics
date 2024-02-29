package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLog *zap.Logger

func init() {
	var err error
	//config := zap.NewProductionConfig()
	//enccoderConfig := zap.NewProductionEncoderConfig()
	config := zap.NewDevelopmentConfig()
	enccoderConfig := zap.NewDevelopmentEncoderConfig()
	enccoderConfig.TimeKey = "timestamp"
	enccoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05")
	enccoderConfig.CallerKey = ""     // to hide caller line key info
	enccoderConfig.StacktraceKey = "" // to hide stacktrace info
	config.EncoderConfig = enccoderConfig

	zapLog, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		//"srv: cannot initialize zap logger"
		panic(err)
	}
}

func Info(message string, fields ...zap.Field) {
	zapLog.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	zapLog.Debug(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	zapLog.Error(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	zapLog.Fatal(message, fields...)
}

func Sync() {
	zapLog.Sync()
}
