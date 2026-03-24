package app

import (
	"context"
	"insync/internal/transport/grpc/client"
	"log/slog"
	"time"
)

func startGrpcClient(
	certPath string,
	keyPath string,
	caCertPath string,

	networkType string,
	serverAddress string,
	serviceName string,
	serverName string,

	resolverScheme string,

	mdnsResolverIfaces []string,

	loadBalancingPolicy string,
	shouldUseHealthCheck bool,

	// gRPC options
	maxAttempts int,
	initialBackoff int,
	maxBackoff int,
	backoffMultiplier float64,

	// Connection options
	baseDelay int,
	multiplier float64,
	maxDelay int,
	jitter float64,
	minConnectTimeout int,

	ctx context.Context,

	chunkSizeInBytes int,

	logger *slog.Logger,

) error {

	rpcRetryPolicy := client.RpcRetryPolicy{
		MaxAttempts:          maxAttempts,
		InitialBackoff:       time.Duration(initialBackoff) * time.Second,
		MaxBackoff:           time.Duration(maxBackoff) * time.Second,
		BackoffMultiplier:    backoffMultiplier,
		RetryableStatusCodes: []string{"UNAVAILABLE", "RESOURCE_EXHAUSTED"},
	}

	connectionConfig := client.ConnectionConfig{
		BaseDelay:         time.Duration(baseDelay) * time.Second,
		Multiplier:        multiplier,
		MaxDelay:          time.Duration(maxDelay) * time.Second,
		Jitter:            jitter,
		MinConnectTimeout: time.Duration(minConnectTimeout) * time.Second,
	}

	grpcClientOptions := client.GrpcClientOptions{
		CertPath:   certPath,
		KeyPath:    keyPath,
		CaCertPath: caCertPath,

		NetworkType:   networkType,
		ServerAddress: serverAddress,
		ServiceName:   serviceName,
		ServerName:    serverName,

		ResolverScheme: resolverScheme,

		MdnsResolverIfaces: mdnsResolverIfaces,

		LoadBalancingPolicy:  loadBalancingPolicy,
		ShouldUseHealthCheck: shouldUseHealthCheck,

		RpcRetryPolicy:   &rpcRetryPolicy,
		ConnectionConfig: &connectionConfig,

		Ctx: ctx,

		ChunkSizeInBytes: chunkSizeInBytes,

		Logger: logger,
	}

	grpcClient := client.NewGrpcClient(grpcClientOptions)
	if err := grpcClient.Start(); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		err := grpcClient.Stop()
		if err != nil {
			panic("Не удалось остановить grpcClient")
		}
	}()

	return nil
}
