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

	ctx context.Context,
	logger *slog.Logger,
) {
	grpcServerOpts := serv.GrpcServerOptions{
		CertPath:   certPath,
		KeyPath:    keyPath,
		CaCertPath: caCertPath,

		NetworkType: networkType,
		Address:     address,

		Ctx: ctx,

		Logger: logger,
	}

	grpcServer := serv.NewGrpcServer(grpcServerOpts)
	go func() {
		err := grpcServer.Start()
		if err != nil {
			panic("Не удалось запустить grpcServer")
		}
	}()

	go func() {
		<-ctx.Done()
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), grpcServerGracefulStopTimeoutSeconds*time.Second)
		defer timeoutCancel()

		err := grpcServer.Stop(timeoutCtx)
		if err != nil {
			logger.Warn("Не удалось коректно остановить grpcServer")
		}
	}()
}
