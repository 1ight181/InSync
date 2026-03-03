package models

import (
	conferr "insync/internal/config/errors"
	"strconv"
)

// Реализует интерфейс ConfigModel
type ServerConfig struct {
	Ip          string `mapstructure:"ip"`
	Port        int    `mapstructure:"port"`
	NetworkType string `mapstructure:"network_type"`
}

func (sc *ServerConfig) Validate() error {
	if sc.Port < 0 || sc.Port > 65535 {
		return conferr.ErrServerPortIsInvalid
	}
	if sc.NetworkType == "" {
		return conferr.ErrServerNetworkTypeIsEmpty
	}

	return nil
}

func (sc *ServerConfig) GetServerAddress() string {
	return sc.Ip + ":" + strconv.Itoa(sc.Port)
}
