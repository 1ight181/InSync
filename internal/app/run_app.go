package app

import (
	"context"
	"fmt"
	"insync/internal/infrastructure/filesys"
	"insync/internal/infrastructure/root"
	rootrepo "insync/internal/repository/sqlite/root"
	baseusecase "insync/internal/usecase/base"
	"time"
)

func RunApp() {
	var cleanups cleanupStack

	config, err := createConfig()
	if err != nil {
		panic(fmt.Sprintf("Не удалось загрузить и проверить конфиг: %v", err))
	}

	logger, err := createLogger(config.LoggerConfig)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать логгер: %v", err))
	}

	db, err := createDb(config.DbConfig, logger, &cleanups)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать БД: %v", err))
	}

	fileSystem := filesys.NewFileSystem()

	clientLogger := logger.With(moduleAtrributeName, grpcClientModuleName)
	connectionManagerLogger := logger.With(moduleAtrributeName, connectionManagerModuleName)
	resolverLogger := logger.With(moduleAtrributeName, mDnsGrpcResolverModuleName)

	connectionManager, err := createConnectionManager(
		connectionManagerLogger,
		clientLogger,
		resolverLogger,
		config.MDnsBrowserConfig.GetInterfaces(),
		config.ClientConfig,
	)

	deviceIdProvider, err := createDeviceIdProvider(
		clientLogger,
		connectionManager,
		config.DeviceIdResolverConfig,
		fileSystem,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать DeviceIdProvider: %v", err))
	}

	rootRepo := rootrepo.NewRootRepository(db)

	rootResolver, err := root.NewRootResolver(rootRepo)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать RootResolver: %v", err))
	}

	hashManagerLogger := logger.With(moduleAtrributeName, hashManagerModuleName)
	fileManagerLogger := logger.With(moduleAtrributeName, fileManagerModuleName)

	fileManager, err := createFileManager(
		hashManagerLogger,
		fileManagerLogger,
		db,
		fileSystem,
		rootResolver,
		config.HashCacheConfig.HashCacheEntryExpireUnixTime,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать fileManager: %v", err))
	}

	baseSnapshotManager, err := createBaseSnapshotManager(
		deviceIdProvider,
		db,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать baseSnapshotManager: %v", err))
	}

	localSnapshotProvider, err := createLocalSnapshotProvider(
		fileManager,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать localSnapshotProvider: %v", err))
	}

	serverLogger := logger.With(moduleAtrributeName, grpcServerModuleName)

	baseUseCaseOpts := baseusecase.BaseSnapshotUseCaseOpts{
		BaseSnapshotManager:   baseSnapshotManager,
		LocalSnapshotProvider: localSnapshotProvider,
	}

	baseUseCase, err := baseusecase.NewBaseSnapshotUseCase(baseUseCaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать baseUseCase: %v", err))
	}

	grpcServer, err := startGrpcServer(
		serverLogger,
		&config.ServerConfig,
		fileManager,
		localSnapshotProvider,
		baseSnapshotManager,
		&cleanups,
		baseUseCase,
	)

	if err != nil {
		panic(fmt.Sprintf("Не удалось создать GrpcServer: %v", err))
	}

	if err := grpcServer.Start(); err != nil {
		panic(fmt.Sprintf("Не удалось запустить GrpcServer: %v", err))
	}

	mDnsServerLogger := logger.With(moduleAtrributeName, mDnsServerModuleName)

	mDnsServer, err := createMDnsServer(
		mDnsServerLogger,
		config.MDnsServerConfig,
		deviceIdProvider,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать mDNS сервер: %v", err))
	}

	if err := mDnsServer.Start(); err != nil {
		panic(fmt.Sprintf("Не удалось запустить mDNS сервер: %v", err))
	}

	mDnsBrowserLogger := logger.With(moduleAtrributeName, mDnsBrowserModuleName)
	cliLogger := logger.With(moduleAtrributeName, cliModuleName)

	cliInstance, err := createCLi(
		cliLogger,
		mDnsBrowserLogger,
		config.MDnsBrowserConfig,
		baseSnapshotManager,
		fileManager,
		connectionManager,
		deviceIdProvider,
		localSnapshotProvider,
		rootResolver,
		db,
		baseUseCase,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать CLI: %v", err))
	}

	cliInstance.Start()
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := cleanups.Run()

	select {
	case <-timeoutCtx.Done():
	case <-done:
	}

}
