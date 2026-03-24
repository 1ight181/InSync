package app

import (
	"context"
	"insync/internal/transport/grpc/client"
	"log/slog"
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

	ctx context.Context,

	chunkSizeInBytes int,

	logger *slog.Logger,

) error {

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

		LoadBalancingPolicy: loadBalancingPolicy,

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
