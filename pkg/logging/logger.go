package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/************
 ** Logger **
 ************/

var Logger *zap.Logger = initLogger()

// initLogger initializes the logger
func initLogger() *zap.Logger {
	// Create a new logger with our custom configuration
	config := zap.NewDevelopmentConfig()
	// Add color on stdout/stderr
	// config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	config.DisableStacktrace = true // Disable stacktrace after WARN or ERROR
	// Save in a log file
	config.OutputPaths = []string{"app.log"}
	config.ErrorOutputPaths = []string{"app.log"}

	logger, _ := config.Build()
	logger.Debug("Logger initialized")
	return logger
}
