package models

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	AllowedLogLevels = map[string]struct{}{"DEBUG": {}, "INFO": {}, "WARN": {}, "ERROR": {}}
)

type LoggerConfig struct {
	ShouldLogToFile  bool   `mapstructure:"should_log_to_file"`
	LogFileDirectory string `mapstructure:"log_file_directory"`
	LogFileName      string `mapstructure:"log_file_name"`
	LogFileExtension string `mapstructure:"log_file_extension"`
	LogLevel         string `mapstructure:"log_level"`
}

func (lc *LoggerConfig) Validate() error {
	if lc.LogFileDirectory == "" {
		return ErrLogFileDirectoryIsEmpty
	}
	if lc.LogFileName == "" {
		return ErrLogFileNameIsEmpty
	}
	if lc.LogFileExtension == "" {
		return ErrLogFileExtensionIsEmpty
	}

	logFilePath := lc.GetLogFilePath()
	if _, err := os.Stat(logFilePath); errors.Is(err, os.ErrNotExist) {
		return &LogFileDoesNotExistError{Err: err}
	}

	lc.LogLevel = strings.ToUpper(lc.LogLevel)

	if _, ok := AllowedLogLevels[lc.LogLevel]; !ok {
		return ErrInvalidLogLevel
	}

	return nil
}

func (lc *LoggerConfig) GetLogFilePath() string {
	normalizedDir := filepath.FromSlash(lc.LogFileDirectory)
	return filepath.Join(normalizedDir, lc.LogFileName+"."+lc.LogFileExtension)
}
