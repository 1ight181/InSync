package app

import (
	loghandler "insync/internal/infrastructure/logger"
	"log/slog"
	"os"
)

func createLogger(
	shouldLogToFile bool,
	logFilePath string,
	logLevel string,
) (*slog.Logger, error) {
	var slogLogLevel slog.Level
	switch logLevel {
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

	if shouldLogToFile {
		logFile, err := os.Create(logFilePath)
		if err != nil {
			return nil, err
		}
		fileOutputHandler := slog.NewJSONHandler(logFile, nil)

		logHandlers = append(logHandlers, fileOutputHandler)
	}

	applicationLogger := slog.New(
		loghandler.NewCombinedHandler(logHandlers),
	)

	return applicationLogger, nil
}
