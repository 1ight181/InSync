package app

import (
	"context"
	clt "insync/internal/client"
	"log/slog"
)

func startGrpcClient(
	clientCertPath string,
	clientKeyPath string,
	caCertPath string,

	clientNetworkType string,
	clientAddress string,

	ctx context.Context,

	chunkSizeInBytes int,

	logger *slog.Logger,

) error {
	clientOpts := clt.GrpcClientOptions{
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		CaCertPath:     caCertPath,

		ClientNetworkType: clientNetworkType,
		ClientAddress:     clientAddress,

		Ctx: ctx,

		ChunkSizeInBytes: chunkSizeInBytes,

		Logger: logger,
	}

	grpcClient := clt.NewGrpcClient(clientOpts)
	err := grpcClient.Start()
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
