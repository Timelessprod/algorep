package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger = initLogger()

func initLogger() *zap.Logger {
	// Create a new logger with our custom configuration
	config := zap.NewDevelopmentConfig()
	// config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add color on stdout/stderr
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	config.DisableStacktrace = true // Disable stacktrace after WARN or ERROR
	config.OutputPaths = []string{"app.log"}
	config.ErrorOutputPaths = []string{"app.log"}

	logger, _ := config.Build()

	logger.Debug("Logger initialized")
	return logger
}
