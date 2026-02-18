package models

import (
	"errors"
	"os"
	"path/filepath"

	conferr "insync/internal/config/errors"
)

type TlsConfig struct {
	ServerCertDir       string
	ServerCertFilename  string
	ServerCertExtension string

	ServerKeyDir       string
	ServerKeyFilename  string
	ServerKeyExtension string

	ClientCertDir       string
	ClientCertFilename  string
	ClientCertExtension string

	ClientKeyDir       string
	ClientKeyFilename  string
	ClientKeyExtension string

	CaCertDir       string
	CaCertFilename  string
	CaCertExtension string
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
	return filepath.Join(cc.ServerCertDir, cc.ServerCertFilename+cc.ServerCertExtension)
}

func (cc *TlsConfig) GetServerKeyPath() string {
	return filepath.Join(cc.ServerKeyDir, cc.ServerKeyFilename+cc.ServerKeyExtension)
}

func (cc *TlsConfig) GetClientCertPath() string {
	return filepath.Join(cc.ClientCertDir, cc.ClientCertFilename+cc.ClientCertExtension)
}

func (cc *TlsConfig) GetClientKeyPath() string {
	return filepath.Join(cc.ClientKeyDir, cc.ClientKeyFilename+cc.ClientKeyExtension)
}

func (cc *TlsConfig) GetCaCertPath() string {
	return filepath.Join(cc.CaCertDir, cc.CaCertFilename+cc.CaCertExtension)
}
