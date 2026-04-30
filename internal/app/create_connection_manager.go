package app

import (
	clt "insync/internal/infrastructure/client"
	mdnsresolver "insync/internal/infrastructure/client/resolver"
	"insync/internal/infrastructure/config/models"
	conn "insync/internal/infrastructure/connection"
	"insync/internal/infrastructure/holder"
	"log/slog"
	"time"

	"google.golang.org/grpc/resolver"
)

func createConnectionManager(
	connectionManagerLogger *slog.Logger,
	clientLogger *slog.Logger,
	resolverLogger *slog.Logger,
	resolverIfaces []string,
	clientConfig models.ClientConfig,
) (*conn.ConnectionManager, error) {
	builderOptions := mdnsresolver.BuilderOptions{
		ResolverIfaces:    resolverIfaces,
		ShouldResolveIpv6: true,
		ShouldReportError: true,
		Logger:            resolverLogger,
	}

	mDnsResolverBuilder := mdnsresolver.NewBuilder(
		builderOptions,
	)

	// В GrpcConf не передается ServerAddress, т.к. он добавляется в ConnectionManager при подключении к конкретному узлу
	// Данная конфигурация является основой для подключения к любому узлу
	grpcClientConf := clt.GrpcConf{
		CertPath: clientConfig.TlsConfig.GetClientCertPath(),
		KeyPath:  clientConfig.TlsConfig.GetClientKeyPath(),

		CaCertPath: clientConfig.TlsConfig.GetCaCertPath(),

		ServerNetworkType: clientConfig.ServerNetworkType,
		ServerServiceName: clientConfig.ServerServiceName,

		ResolverScheme: clientConfig.ResolverScheme,

		LoadBalancingPolicy:  clientConfig.LoadBalancingPolicy,
		ShouldUseHealthCheck: clientConfig.ShouldUseHealthCheck,

		RpcTimeout: time.Duration(clientConfig.RpcTimeout) * time.Second,

		RetryPolicy: &clt.RpcRetryPolicy{
			MaxAttempts:          clientConfig.RpcRetryPolicy.MaxAttempts,
			InitialBackoff:       time.Duration(clientConfig.RpcRetryPolicy.InitialBackoffSeconds) * time.Second,
			MaxBackoff:           time.Duration(clientConfig.RpcRetryPolicy.MaxBackoffSeconds) * time.Second,
			BackoffMultiplier:    clientConfig.RpcRetryPolicy.BackoffMultiplier,
			RetryableStatusCodes: clientConfig.RpcRetryPolicy.RetryableStatusCodes,
		},

		ConnectionConfig: &clt.ConnectionConfig{
			BaseDelay:         time.Duration(clientConfig.ConnectionConfig.BaseDelaySeconds) * time.Second,
			Multiplier:        clientConfig.ConnectionConfig.Multiplier,
			MaxDelay:          time.Duration(clientConfig.ConnectionConfig.MaxDelaySeconds) * time.Second,
			Jitter:            clientConfig.ConnectionConfig.Jitter,
			MinConnectTimeout: time.Duration(clientConfig.ConnectionConfig.MinConnectTimeoutSeconds) * time.Second,
		},

		ChunkSizeInBytes: clientConfig.ChunkSizeInBytes,
		Resolvers:        []resolver.Builder{mDnsResolverBuilder},
	}

	clientHolder := holder.NewClientHolder()

	connectionManagerOpts := conn.ConnectionManagerOptions{
		BaseGrpcConf:      &grpcClientConf,
		ClientHolder:      clientHolder,
		ServerServiceType: clientConfig.MdnsServerServiceType,
		ServerNamePrefix:  clientConfig.MdnsServerNamePrefix,
		GrpcClientLogger:  clientLogger,
		Logger:            connectionManagerLogger,
	}

	connectionManager, err := conn.NewConnectionManager(connectionManagerOpts)
	if err != nil {
		return nil, err
	}

	return connectionManager, nil
}
