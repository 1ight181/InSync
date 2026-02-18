package models

import (
	"errors"
	conferr "insync/internal/config/errors"
	"os"
	"path/filepath"
)

// Реализует интерфейс ConfigModel
type ServerConfig struct {
	CertDir       string
	CertFilename  string
	CertExtension string

	Ip          string
	Port        int
	NetworkType string
}

func (sc *ServerConfig) Validate() error {
	if sc.CertDir == "" {
		return conferr.ErrCertDirIsEmpty
	}
	if sc.CertFilename == "" {
		return conferr.ErrCertFilenameIsEmpty
	}
	if sc.CertExtension == "" {
		return conferr.ErrCertExtensionIsEmpty
	}

	filePath := sc.GetCertPath()
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return &conferr.CertFileDoesNotExistError{FilePath: filePath}
	}

	if sc.Ip == "" {
		return conferr.ErrServerIpIsEmpty
	}
	if sc.Port < 0 || sc.Port > 65535 {
		return conferr.ErrServerPortIsInvalid
	}
	if sc.NetworkType == "" {
		return conferr.ErrServerNetworkTypeIsEmpty
	}

	return nil
}

func (sc *ServerConfig) GetCertPath() string {
	return filepath.Join(sc.CertDir, sc.CertFilename+sc.CertExtension)
}
