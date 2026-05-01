package app

import (
	"context"
	"insync/internal/infrastructure/base"
	"insync/internal/infrastructure/config/models"
	"insync/internal/infrastructure/filemanager"
	local "insync/internal/infrastructure/local"
	"insync/internal/transport/grpc/server"
	baseusecase "insync/internal/usecase/base"
	fileusecase "insync/internal/usecase/file"
	"log/slog"
	"time"
)

func startGrpcServer(
	serverLogger *slog.Logger,
	serverConfig *models.ServerConfig,
	fileManager *filemanager.FileManager,
	localSnapshotProvider *local.LocalSnapshotProvider,
	baseSnapshotManager *base.BaseSnapshotManager,
	cleanups *cleanupStack,
	baseUseCase *baseusecase.BaseSnapshotUseCase,
) (*server.GrpcServer, error) {
	fileUseCaseOpts := fileusecase.FileUseCaseOptions{
		FileManager:          fileManager,
		BaseSnapshotProvider: baseSnapshotManager,
	}

	fileUseCase, err := fileusecase.NewFileUseCase(fileUseCaseOpts)
	if err != nil {
		return nil, err
	}

	grpcServerOpts := server.GrpcServerOptions{
		FileUseCase: fileUseCase,
		BaseUseCase: baseUseCase,

		CertPath:   serverConfig.TlsConfig.GetServerCertPath(),
		KeyPath:    serverConfig.TlsConfig.GetServerKeyPath(),
		CaCertPath: serverConfig.TlsConfig.GetCaCertPath(),

		NetworkType: serverConfig.NetworkType,
		Address:     serverConfig.GetAddress(),
		ServiceName: serverConfig.ServiceName,

		ChunkSizeInBytes: serverConfig.ChunkSizeInBytes,

		ShouldStartHealthServer: true,

		Logger: serverLogger,
	}

	grpcServer, err := server.NewGrpcServer(grpcServerOpts)
	if err != nil {
		return nil, err
	}

	cleanups.Add(func() {
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), grpcServerGracefulStopTimeoutSeconds*time.Second)
		defer timeoutCancel()

		err := grpcServer.Stop(timeoutCtx)
		if err != nil {
			serverLogger.Warn("Не удалось коректно остановить grpcServer")
		}
	})

	return grpcServer, nil
}
