package models

import (
	"errors"
	conferr "insync/internal/config/errors"
	"os"
	"path/filepath"
)

type LoggerConfig struct {
	ShouldLogToFile  bool   `mapstructure:"should_log_to_file"`
	LogFileDirectory string `mapstructure:"log_file_directory"`
	LogFileName      string `mapstructure:"log_file_name"`
	LogFileExtension string `mapstructure:"log_file_extension"`
	LogLevel         string `mapstructure:"log_level"`
}

func (lc LoggerConfig) Validate() error {
	if lc.LogFileDirectory == "" {
		return conferr.ErrLogFileDirectoryIsEmpty
	}
	if lc.LogFileName == "" {
		return conferr.ErrLogFileNameIsEmpty
	}
	if lc.LogFileExtension == "" {
		return conferr.ErrLogFileExtensionIsEmpty
	}

	logFilePath := lc.GetLogFilePath()
	if _, err := os.Stat(logFilePath); errors.Is(err, os.ErrNotExist) {
		return &conferr.LogFileDoesNotExistError{Err: err}
	}

	if lc.LogLevel == "" {
		return conferr.ErrLogLevelIsEmpty
	}

	return nil
}

func (lc LoggerConfig) GetLogFilePath() string {
	normalizedDir := filepath.FromSlash(lc.LogFileDirectory)
	return filepath.Join(normalizedDir, lc.LogFileName+"."+lc.LogFileExtension)
}
