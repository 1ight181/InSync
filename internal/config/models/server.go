package models

import (
	conferr "insync/internal/config/errors"
)

// Реализует интерфейс ConfigModel
type ServerConfig struct {
	Ip          string
	Port        int
	NetworkType string
}

func (sc *ServerConfig) Validate() error {
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
