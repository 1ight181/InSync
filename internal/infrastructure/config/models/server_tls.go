package models

import (
	"errors"
	"os"
	"path/filepath"
)

type ServerTlsConfig struct {
	CommonTlsConfig     `mapstructure:",squash"`
	ServerCertDir       string `mapstructure:"server_cert_dir"`
	ServerCertFilename  string `mapstructure:"server_cert_filename"`
	ServerCertExtension string `mapstructure:"server_cert_extension"`

	ServerKeyDir       string `mapstructure:"server_key_dir"`
	ServerKeyFilename  string `mapstructure:"server_key_filename"`
	ServerKeyExtension string `mapstructure:"server_key_extension"`
}

func (cc *ServerTlsConfig) Validate() error {
	if err := cc.validateServerCert(); err != nil {
		return err
	}

	if err := cc.validateServerKey(); err != nil {
		return err
	}

	return nil
}

func (cc *ServerTlsConfig) validateServerCert() error {
	if cc.ServerCertDir == "" {
		return ErrServerCertDirIsEmpty
	}
	if cc.ServerCertFilename == "" {
		return ErrServerCertFilenameIsEmpty
	}
	if cc.ServerCertExtension == "" {
		return ErrServerCertExtensionIsEmpty
	}

	serverSertPath := cc.GetServerCertPath()
	if _, err := os.Stat(serverSertPath); errors.Is(err, os.ErrNotExist) {
		return &ServerCertFileDoesNotExistError{FilePath: serverSertPath}
	}

	return nil
}

func (cc *ServerTlsConfig) validateServerKey() error {
	if cc.ServerKeyDir == "" {
		return ErrServerKeyDirIsEmpty
	}
	if cc.ServerKeyFilename == "" {
		return ErrServerKeyFilenameIsEmpty
	}
	if cc.ServerKeyExtension == "" {
		return ErrServerKeyExtensionIsEmpty
	}

	serverKeyPath := cc.GetServerKeyPath()
	if _, err := os.Stat(serverKeyPath); errors.Is(err, os.ErrNotExist) {
		return &ServerKeyFileDoesNotExistError{FilePath: serverKeyPath}
	}

	return nil
}

func (cc *ServerTlsConfig) GetServerCertPath() string {
	normalizedDir := filepath.FromSlash(cc.ServerCertDir)
	return filepath.Join(normalizedDir, cc.ServerCertFilename+cc.ServerCertExtension)
}

func (cc *ServerTlsConfig) GetServerKeyPath() string {
	normalizedDir := filepath.FromSlash(cc.ServerKeyDir)
	return filepath.Join(normalizedDir, cc.ServerKeyFilename+cc.ServerKeyExtension)
}
