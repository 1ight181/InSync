package models

import (
	"fmt"
	"strconv"
	"strings"
)

type ClientConfig struct {
	// passthrough/mdns
	ResolverScheme string `mapstructure:"resolver_scheme"`

	// для passthrough
	ServerAddress string `mapstructure:"server_address"`

	//для mdns
	ServerServiceName string `mapstructure:"server_service_name"`
	ServerInterfaces  string `mapstructure:"server_interfaces"`

	// общие
	ServerPort           int    `mapstructure:"server_port"`
	ServerNetworkType    string `mapstructure:"server_network_type"`
	LoadBalancingPolicy  string `mapstructure:"load_balancing_policy"`
	ShouldUseHealthCheck bool   `mapstructure:"should_use_health_check"`

	RpcTimeout int `mapstructure:"rpc_timeout"`

	RpcRetryPolicy   RpcRetryPolicy   `mapstructure:"rpc_retry_policy"`
	ConnectionConfig ConnectionConfig `mapstructure:"connection_config"`

	ChunkSizeInBytes int `mapstructure:"chunk_size_in_bytes"`
}

func (cc *ClientConfig) Validate() error {
	if cc.GetServerAddress() == "" && cc.ResolverScheme == "passthrough" {
		return ErrAddressToConnectIsEmpty
	}
	if cc.ServerPort < 0 || cc.ServerPort > 65535 {
		return ErrPortToConnectIsInvalid
	}
	if cc.ServerNetworkType == "" {
		return ErrNetworkToConnectTypeIsEmpty
	}
	if cc.ResolverScheme == "" {
		return ErrClientResolverSchemeIsEmpty
	}
	if cc.ChunkSizeInBytes <= 0 {
		return ErrChunkSizeIsInvalid
	}
	if cc.LoadBalancingPolicy == "" {
		return ErrLoadBalancingPolicyIsEmpty
	}

	if cc.RpcTimeout <= 0 {
		return ErrRpcTimeoutIsInvalid
	}

	if err := cc.RpcRetryPolicy.Validate(); err != nil {
		return err
	}

	if err := cc.ConnectionConfig.Validate(); err != nil {
		return err
	}

	if cc.ResolverScheme == "mdns" && cc.ServerServiceName == "" {
		if cc.ServerAddress == "" {
			return ErrAddressToConnectIsEmpty
		}
		if cc.ServerAddress == "" {
			return ErrAddressToConnectIsEmpty
		}
	}

	return nil
}

func (cc *ClientConfig) GetServerAddress() string {
	if cc.ServerAddress == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s", cc.ServerAddress, strconv.Itoa(cc.ServerPort))
}

func (cc *ClientConfig) GetInterfaces() []string {
	return strings.Split(cc.ServerInterfaces, ";")
}
