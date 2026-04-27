package models

import (
	"fmt"
	"strconv"
)

type ClientConfig struct {
	// passthrough/mdns
	ResolverScheme string `mapstructure:"resolver_scheme"`

	// для passthrough
	ServerIp   string `mapstructure:"server_ip"`
	ServerPort int    `mapstructure:"server_port"`
	// общие
	ServerServiceName    string `mapstructure:"server_service_name"`
	ServerNetworkType    string `mapstructure:"server_network_type"`
	LoadBalancingPolicy  string `mapstructure:"load_balancing_policy"`
	ShouldUseHealthCheck bool   `mapstructure:"should_use_health_check"`

	RpcTimeout int `mapstructure:"rpc_timeout"`

	RpcRetryPolicy   RpcRetryPolicy   `mapstructure:"rpc_retry_policy"`
	ConnectionConfig ConnectionConfig `mapstructure:"connection_config"`

	ChunkSizeInBytes int `mapstructure:"chunk_size_in_bytes"`

	TlsConfig ClientTlsConfig `mapstructure:"tls"`
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

	return nil
}

func (cc *ClientConfig) GetServerAddress() string {
	if cc.ServerIp == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s", cc.ServerIp, strconv.Itoa(cc.ServerPort))
}
