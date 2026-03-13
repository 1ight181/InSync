package models

import (
	"fmt"
	conferr "insync/internal/config/errors"
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
		return conferr.ErrClientPortIsInvalid
	}
	if cc.ServerNetworkType == "" {
		return conferr.ErrClientNetworkTypeIsEmpty
	}
	if cc.ResolverScheme == "" {
		return conferr.ErrClientResolverSchemeIsEmpty
	}
	if cc.ChunkSizeInBytes <= 0 {
		return conferr.ErrChunkSizeIsInvalid
	}

	return nil
}

func (cc *ClientConfig) GetServerAddress() string {
	if cc.ServerIp == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s", cc.ServerIp, strconv.Itoa(cc.ServerPort))
}
