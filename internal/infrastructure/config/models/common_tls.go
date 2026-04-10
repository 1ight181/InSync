package models

import (
	"errors"
	"os"
	"path/filepath"
)

type CommonTlsConfig struct {
	CaCertDir       string `mapstructure:"ca_cert_dir"`
	CaCertFilename  string `mapstructure:"ca_cert_filename"`
	CaCertExtension string `mapstructure:"ca_cert_extension"`
}

func (cc *CommonTlsConfig) Validate() error {
	if err := cc.validateCaCert(); err != nil {
		return err
	}

	return nil
}

func (cc *CommonTlsConfig) validateCaCert() error {
	if cc.CaCertDir == "" {
		return ErrCaCertDirIsEmpty
	}
	if cc.CaCertFilename == "" {
		return ErrCaCertFilenameIsEmpty
	}
	if cc.CaCertExtension == "" {
		return ErrCaCertExtensionIsEmpty
	}

	caCertPath := cc.GetCaCertPath()
	if _, err := os.Stat(caCertPath); errors.Is(err, os.ErrNotExist) {
		return &CaCertFileDoesNotExistError{FilePath: caCertPath}
	}
	return nil
}

func (cc *CommonTlsConfig) GetCaCertPath() string {
	normalizedDir := filepath.FromSlash(cc.CaCertDir)
	return filepath.Join(normalizedDir, cc.CaCertFilename+cc.CaCertExtension)
}
