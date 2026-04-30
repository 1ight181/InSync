package app

import (
	"fmt"
	config "insync/internal/infrastructure/config/models"
	loghandler "insync/internal/infrastructure/logger"
	"log/slog"
	"os"
)

func createLogger(
	loggerConfig config.LoggerConfig,
) (*slog.Logger, error) {
	logDir := loggerConfig.LogFileDirectory
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("Не удалось создать директорию для логов: %v", err))
	}

	var slogLogLevel slog.Level
	switch loggerConfig.LogLevel {
	case "DEBUG":
		slogLogLevel = slog.LevelDebug
	case "INFO":
		slogLogLevel = slog.LevelInfo
	case "WARN":
		slogLogLevel = slog.LevelWarn
	case "ERROR":
		slogLogLevel = slog.LevelError
	default:
		slogLogLevel = slog.LevelInfo
	}

	var logHandlers []slog.Handler
	terminalOutputHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLogLevel,
	})

	logHandlers = append(logHandlers, terminalOutputHandler)

	if loggerConfig.ShouldLogToFile {
		logFile, err := os.OpenFile(loggerConfig.GetLogFilePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		fileOutputHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: slogLogLevel,
		})

		logHandlers = append(logHandlers, fileOutputHandler)
	}

	applicationLogger := slog.New(
		loghandler.NewCombinedHandler(logHandlers),
	)

	return applicationLogger, nil
}
