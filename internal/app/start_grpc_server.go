package app

import (
	"context"
	serv "insync/internal/server"
	"log/slog"
	"time"
)

const (
	grpcServerGracefulStopTimeoutSeconds = exitTimeoutSeconds - 1
)

func startGrpcServer(
	serverCertPath string,
	serverKeyPath string,
	caCertPath string,

	serverNetworkType string,
	serverAddress string,

	ctx context.Context,
	logger *slog.Logger,
) {
	grpcServerOpts := serv.GrpcServerOptions{
		ServerCertPath: serverCertPath,
		ServerKeyPath:  serverKeyPath,
		CaCertPath:     caCertPath,

		ServerNetworkType: serverNetworkType,
		ServerAddress:     serverAddress,

		Ctx:    ctx,
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
