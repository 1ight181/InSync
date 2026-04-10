package models

import (
	"strconv"
)

// Реализует интерфейс ConfigModel
type ServerConfig struct {
	Address     string `mapstructure:"address"`
	Port        int    `mapstructure:"port"`
	NetworkType string `mapstructure:"network_type"`
	ServiceName string `mapstructure:"service_name"`

	TlsConfig ServerTlsConfig `mapstructure:"tls"`
}

func (sc *ServerConfig) Validate() error {
	if sc.Port < 0 || sc.Port > 65535 {
		return ErrServerPortIsInvalid
	}
	if sc.NetworkType == "" {
		return ErrServerNetworkTypeIsEmpty
	}

	return nil
}

func (sc *ServerConfig) GetAddress() string {
	return sc.Address + ":" + strconv.Itoa(sc.Port)
}
