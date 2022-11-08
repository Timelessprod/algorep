package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initLogger() *zap.Logger {
	// Create a new logger with our custom configuration
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.DisableStacktrace = true // Disable stacktrace after WARN or ERROR

	logger, _ := config.Build()

	logger.Info("Logger initialized")
	return logger
}

func main() {
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Hello World")
}
