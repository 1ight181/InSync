package models

import (
	"errors"
	"os"
	"path/filepath"

	conferr "insync/internal/config/errors"
)

type TlsConfig struct {
	ServerCertDir       string `mapstructure:"server_cert_dir"`
	ServerCertFilename  string `mapstructure:"server_cert_filename"`
	ServerCertExtension string `mapstructure:"server_cert_extension"`

	ServerKeyDir       string `mapstructure:"server_key_dir"`
	ServerKeyFilename  string `mapstructure:"server_key_filename"`
	ServerKeyExtension string `mapstructure:"server_key_extension"`

	ClientCertDir       string `mapstructure:"client_cert_dir"`
	ClientCertFilename  string `mapstructure:"client_cert_filename"`
	ClientCertExtension string `mapstructure:"client_cert_extension"`

	ClientKeyDir       string `mapstructure:"client_key_dir"`
	ClientKeyFilename  string `mapstructure:"client_key_filename"`
	ClientKeyExtension string `mapstructure:"client_key_extension"`

	CaCertDir       string `mapstructure:"ca_cert_dir"`
	CaCertFilename  string `mapstructure:"ca_cert_filename"`
	CaCertExtension string `mapstructure:"ca_cert_extension"`
}

func (cc *TlsConfig) validateServerCert() error {
	if cc.ServerCertDir == "" {
		return conferr.ErrServerCertDirIsEmpty
	}
	if cc.ServerCertFilename == "" {
		return conferr.ErrServerCertFilenameIsEmpty
	}
	if cc.ServerCertExtension == "" {
		return conferr.ErrServerCertExtensionIsEmpty
	}

	serverSertPath := cc.GetServerCertPath()
	if _, err := os.Stat(serverSertPath); errors.Is(err, os.ErrNotExist) {
		return &conferr.ServerCertFileDoesNotExistError{FilePath: serverSertPath}
	}

	return nil
}

func (cc *TlsConfig) validateServerKey() error {
	if cc.ServerKeyDir == "" {
		return conferr.ErrServerKeyDirIsEmpty
	}
	if cc.ServerKeyFilename == "" {
		return conferr.ErrServerKeyFilenameIsEmpty
	}
	if cc.ServerKeyExtension == "" {
		return conferr.ErrServerKeyExtensionIsEmpty
	}

	serverKeyPath := cc.GetServerKeyPath()
	if _, err := os.Stat(serverKeyPath); errors.Is(err, os.ErrNotExist) {
		return &conferr.ServerKeyFileDoesNotExistError{FilePath: serverKeyPath}
	}

	return nil
}

func (cc *TlsConfig) validateClientCert() error {
	if cc.ClientCertDir == "" {
		return conferr.ErrClientCertDirIsEmpty
	}
	if cc.ClientCertFilename == "" {
		return conferr.ErrClientCertFilenameIsEmpty
	}
	if cc.ClientCertExtension == "" {
		return conferr.ErrClientCertExtensionIsEmpty
	}

	clientCertPath := cc.GetClientCertPath()
	if _, err := os.Stat(clientCertPath); errors.Is(err, os.ErrNotExist) {
		return &conferr.ClientCertFileDoesNotExistError{FilePath: clientCertPath}
	}

	return nil
}

func (cc *TlsConfig) validateClientKey() error {
	if cc.ClientKeyDir == "" {
		return conferr.ErrClientKeyDirIsEmpty
	}
	if cc.ClientKeyFilename == "" {
		return conferr.ErrClientKeyFilenameIsEmpty
	}
	if cc.ClientKeyExtension == "" {
		return conferr.ErrClientKeyExtensionIsEmpty
	}

	clientKeyPath := cc.GetClientKeyPath()
	if _, err := os.Stat(clientKeyPath); errors.Is(err, os.ErrNotExist) {
		return &conferr.ClientKeyFileDoesNotExistError{FilePath: clientKeyPath}
	}

	return nil
}

func (cc *TlsConfig) validateCaCert() error {
	if cc.CaCertDir == "" {
		return conferr.ErrCaCertDirIsEmpty
	}
	if cc.CaCertFilename == "" {
		return conferr.ErrCaCertFilenameIsEmpty
	}
	if cc.CaCertExtension == "" {
		return conferr.ErrCaCertExtensionIsEmpty
	}

	caCertPath := cc.GetCaCertPath()
	if _, err := os.Stat(caCertPath); errors.Is(err, os.ErrNotExist) {
		return &conferr.CaCertFileDoesNotExistError{FilePath: caCertPath}
	}
	return nil
}

func (cc *TlsConfig) Validate() error {
	if err := cc.validateServerCert(); err != nil {
		return err
	}

	if err := cc.validateServerKey(); err != nil {
		return err
	}

	if err := cc.validateClientCert(); err != nil {
		return err
	}

	if err := cc.validateClientKey(); err != nil {
		return err
	}

	if err := cc.validateCaCert(); err != nil {
		return err
	}

	return nil
}

func (cc *TlsConfig) GetServerCertPath() string {
	normalizedDir := filepath.FromSlash(cc.ServerCertDir)
	return filepath.Join(normalizedDir, cc.ServerCertFilename+cc.ServerCertExtension)
}

func (cc *TlsConfig) GetServerKeyPath() string {
	normalizedDir := filepath.FromSlash(cc.ServerKeyDir)
	return filepath.Join(normalizedDir, cc.ServerKeyFilename+cc.ServerKeyExtension)
}

func (cc *TlsConfig) GetClientCertPath() string {
	normalizedDir := filepath.FromSlash(cc.ClientCertDir)
	return filepath.Join(normalizedDir, cc.ClientCertFilename+cc.ClientCertExtension)
}

func (cc *TlsConfig) GetClientKeyPath() string {
	normalizedDir := filepath.FromSlash(cc.ClientKeyDir)
	return filepath.Join(normalizedDir, cc.ClientKeyFilename+cc.ClientKeyExtension)
}

func (cc *TlsConfig) GetCaCertPath() string {
	normalizedDir := filepath.FromSlash(cc.CaCertDir)
	return filepath.Join(normalizedDir, cc.CaCertFilename+cc.CaCertExtension)
}
