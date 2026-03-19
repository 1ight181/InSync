package app

import (
	"context"
	cltmng "insync/internal/transport/grpc/client"
	"log/slog"
)

func startGrpcClient(
	certPath string,
	keyPath string,
	caCertPath string,

	networkType string,
	serverAddresses []string,
	resolverScheme string,

	serviceName string,

	ctx context.Context,

	chunkSizeInBytes int,

	logger *slog.Logger,

) error {
	clientManagerOpts := cltmng.GrpcClientManagerOptions{
		CertPath:   certPath,
		KeyPath:    keyPath,
		CaCertPath: caCertPath,

		NetworkType:     networkType,
		ServerAddresses: serverAddresses,
		ResolverScheme:  resolverScheme,

		ServiceName: serviceName,

		Ctx: ctx,

		ChunkSizeInBytes: chunkSizeInBytes,

		Logger: logger,
	}

	grpcClientManager := cltmng.NewGrpcClientManager(clientManagerOpts)
	grpcClient, err := grpcClientManager.GetClient()
	if err != nil {
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
