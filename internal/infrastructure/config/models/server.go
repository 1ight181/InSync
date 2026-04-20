package models

import (
	"strconv"
)

// Реализует интерфейс ConfigModel
type ServerConfig struct {
	Ip               string `mapstructure:"ip"`
	Port             int    `mapstructure:"port"`
	NetworkType      string `mapstructure:"network_type"`
	ServiceName      string `mapstructure:"service_name"`
	ChunkSizeInBytes int    `mapstructure:"chunk_size_in_bytes"`

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
	return sc.Ip + ":" + strconv.Itoa(sc.Port)
}
