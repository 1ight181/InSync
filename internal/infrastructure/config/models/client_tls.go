package models

import (
	"errors"
	"os"
	"path/filepath"
)

type ClientTlsConfig struct {
	CommonTlsConfig     `mapstructure:",squash"`
	ClientCertDir       string `mapstructure:"client_cert_dir"`
	ClientCertFilename  string `mapstructure:"client_cert_filename"`
	ClientCertExtension string `mapstructure:"client_cert_extension"`

	ClientKeyDir       string `mapstructure:"client_key_dir"`
	ClientKeyFilename  string `mapstructure:"client_key_filename"`
	ClientKeyExtension string `mapstructure:"client_key_extension"`
}

func (cc *ClientTlsConfig) Validate() error {

	if err := cc.validateClientCert(); err != nil {
		return err
	}

	if err := cc.validateClientKey(); err != nil {
		return err
	}

	return nil
}

func (cc *ClientTlsConfig) validateClientCert() error {
	if cc.ClientCertDir == "" {
		return ErrClientCertDirIsEmpty
	}
	if cc.ClientCertFilename == "" {
		return ErrClientCertFilenameIsEmpty
	}
	if cc.ClientCertExtension == "" {
		return ErrClientCertExtensionIsEmpty
	}

	clientCertPath := cc.GetClientCertPath()
	if _, err := os.Stat(clientCertPath); errors.Is(err, os.ErrNotExist) {
		return &ClientCertFileDoesNotExistError{FilePath: clientCertPath}
	}

	return nil
}

func (cc *ClientTlsConfig) validateClientKey() error {
	if cc.ClientKeyDir == "" {
		return ErrClientKeyDirIsEmpty
	}
	if cc.ClientKeyFilename == "" {
		return ErrClientKeyFilenameIsEmpty
	}
	if cc.ClientKeyExtension == "" {
		return ErrClientKeyExtensionIsEmpty
	}

	clientKeyPath := cc.GetClientKeyPath()
	if _, err := os.Stat(clientKeyPath); errors.Is(err, os.ErrNotExist) {
		return &ClientKeyFileDoesNotExistError{FilePath: clientKeyPath}
	}

	return nil
}

func (cc *ClientTlsConfig) GetClientCertPath() string {
	normalizedDir := filepath.FromSlash(cc.ClientCertDir)
	return filepath.Join(normalizedDir, cc.ClientCertFilename+cc.ClientCertExtension)
}

func (cc *ClientTlsConfig) GetClientKeyPath() string {
	normalizedDir := filepath.FromSlash(cc.ClientKeyDir)
	return filepath.Join(normalizedDir, cc.ClientKeyFilename+cc.ClientKeyExtension)
}
