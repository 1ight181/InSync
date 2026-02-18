package models

import (
	conferr "insync/internal/config/errors"
)

// Реализует интерфейс ConfigModel
type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

func (dc *DbConfig) Validate() error {
	if dc.Host == "" {
		return conferr.ErrDbHostIsEmpty
	}
	if dc.Port < 0 || dc.Port > 65535 {
		return conferr.ErrDbPortIsInvalid
	}
	if dc.User == "" {
		return conferr.ErrDbUserIsEmpty
	}
	if dc.Password == "" {
		return conferr.ErrDbPasswordIsEmpty
	}

	return nil
}
