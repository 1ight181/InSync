package app

import (
	"context"
	serv "insync/internal/transport/grpc/server"
	"log/slog"
	"time"
)

const (
	grpcServerGracefulStopTimeoutSeconds = exitTimeoutSeconds - 1
)

func startGrpcServer(
	certPath string,
	keyPath string,
	caCertPath string,

	networkType string,
	address string,
	serviceName string,

	ctx context.Context,
	logger *slog.Logger,
) error {
	grpcServerOpts := serv.GrpcServerOptions{
		CertPath:   certPath,
		KeyPath:    keyPath,
		CaCertPath: caCertPath,

		NetworkType: networkType,
		Address:     address,
		ServiceName: serviceName,

		Ctx: ctx,

		Logger: logger,
	}

	grpcServer := serv.NewGrpcServer(grpcServerOpts)
	err := grpcServer.Start()
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), grpcServerGracefulStopTimeoutSeconds*time.Second)
		defer timeoutCancel()

		err := grpcServer.Stop(timeoutCtx)
		if err != nil {
			logger.Warn("Не удалось коректно остановить grpcServer")
		}
	}()

	return nil
}
