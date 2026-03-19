package models

import (
	"fmt"
	"strconv"
)

type ClientConfig struct {
	ServerIp          string `mapstructure:"server_ip"`
	ServerPort        int    `mapstructure:"server_port"`
	ServerNetworkType string `mapstructure:"server_network_type"`
	ResolverScheme    string `mapstructure:"resolver_scheme"`

	ChunkSizeInBytes int    `mapstructure:"chunk_size_in_bytes"`
	ServiceName      string `mapstructure:"service_name"`
}

func (cc *ClientConfig) Validate() error {
	if cc.ServerPort < 0 || cc.ServerPort > 65535 {
		return ErrClientPortIsInvalid
	}
	if cc.ServerNetworkType == "" {
		return ErrClientNetworkTypeIsEmpty
	}
	if cc.ResolverScheme == "" {
		return ErrClientResolverSchemeIsEmpty
	}
	if cc.ChunkSizeInBytes <= 0 {
		return ErrChunkSizeIsInvalid
	}

	return nil
}

func (cc *ClientConfig) GetServerAddress() string {
	if cc.ServerIp == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s", cc.ServerIp, strconv.Itoa(cc.ServerPort))
}
