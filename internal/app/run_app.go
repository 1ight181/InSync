package app

import (
	"context"
	"fmt"
	"insync/internal/domain"
	clt "insync/internal/infrastructure/client"
	conn "insync/internal/infrastructure/connection"
	deviceidlocal "insync/internal/infrastructure/deviceid/local"
	deviceidcreator "insync/internal/infrastructure/deviceid/local/creator"
	"insync/internal/infrastructure/filemanager"
	"insync/internal/infrastructure/filemanager/hash"
	hashcache "insync/internal/infrastructure/filemanager/hash/cache"
	"insync/internal/infrastructure/filemanager/pathtree"
	"insync/internal/infrastructure/filesys"
	mdnsurl "insync/internal/infrastructure/mdnsurl"
	planner "insync/internal/infrastructure/planner"
	planres "insync/internal/infrastructure/planresolver"
	base "insync/internal/infrastructure/planresolver/base"
	local "insync/internal/infrastructure/planresolver/local"
	remote "insync/internal/infrastructure/planresolver/remote"
	"insync/internal/infrastructure/root"
	cli "insync/internal/presentation/cli"
	repo "insync/internal/repository/sqlite"
	baserepo "insync/internal/repository/sqlite/base"
	hashrepo "insync/internal/repository/sqlite/hash"
	server "insync/internal/transport/grpc/server"
	connusecase "insync/internal/usecase/connect"
	fileusecase "insync/internal/usecase/file"
	nodeusecase "insync/internal/usecase/node"
	rootusecase "insync/internal/usecase/root"
	scanusecase "insync/internal/usecase/scan"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	exitTimeoutSeconds                   = 5
	grpcServerGracefulStopTimeoutSeconds = 3
	moduleAtrributeName                  = "module"
	mDnsServerModuleName                 = "mdns_server"
	mDnsBrowserModuleName                = "mdns_browser"
	grpcServerModuleName                 = "grpc_server"
	grpcClientModuleName                 = "grpc_client"
	connectionManagerModuleName          = "connection_manager"
	cliModuleName                        = "cli"
	fileManagerModuleName                = "file_manager"
	hashManagerModuleName                = "hash_manager"
)

func RunApp() {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	appCtx, cancel := context.WithCancel(context.Background())

	config, err := createConfig(appCtx)
	if err != nil {
		panic(fmt.Sprintf("Не удалось загрузить и проверить конфиг: %v", err))
	}

	loggerConfig := config.LoggerConfig
	logger, err := createLogger(
		loggerConfig.ShouldLogToFile,
		loggerConfig.GetLogFilePath(),
		loggerConfig.LogLevel,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать логгер: %v", err))
	}

	mDnsServerLogger := logger.With(moduleAtrributeName, mDnsServerModuleName)
	mDnsServerConfig := config.MDnsServerConfig

	if err := startMDnsServer(
		mDnsServerConfig.InstanceName,
		mDnsServerConfig.ServiceType,
		mDnsServerConfig.Domain,
		mDnsServerConfig.Port,
		mDnsServerConfig.GetInterfaces(),
		mDnsServerLogger,
		appCtx,
	); err != nil {
		panic(fmt.Sprintf("Не удалось запустить mDNS сервер: %v", err))
	}

	rootResolver := root.NewRootResolver()
	fileSystem := filesys.NewFileSystem()
	pathTree := pathtree.NewPathTree()

	dbConfig := config.DbConfig
	dbDir := dbConfig.Dir
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		panic(fmt.Sprintf("Не удалось создать директорию для БД: %v", err))
	}
	dsn := dbConfig.GetDsn()

	migrator := repo.NewAutoMigrator()
	dbOpts := repo.SqliteGormOptions{
		Dsn:      dsn,
		Migrator: migrator,
	}

	db := repo.NewGorm(dbOpts)

	hashCacheRepoOpts := hashrepo.HashRepositoryOptions{
		Db: db,
	}

	hashCacheRepo := hashrepo.NewHashRepository(hashCacheRepoOpts)

	hashCacheOpts := hashcache.HashCacheOptions{
		HashCacheRepository: hashCacheRepo,
	}

	hashCache := hashcache.NewHashCache(hashCacheOpts)
	hashCalc := hash.NewHashCalculator()

	hashManagerLogger := logger.With(moduleAtrributeName, hashManagerModuleName)

	hashManagerOpts := hash.HashManagerOptions{
		HashCache:      hashCache,
		HashCalculator: hashCalc,
		PathTreeReader: pathTree,
		Logger:         hashManagerLogger,
		LoggerCtx:      appCtx,
	}

	hashManager := hash.NewHashManager(hashManagerOpts)

	fileManagerLogger := logger.With(moduleAtrributeName, fileManagerModuleName)

	deviceIdResolverConfig := config.DeviceIdResolverConfig
	deviceIdDir := deviceIdResolverConfig.DeviceIdDir
	if err := os.MkdirAll(deviceIdDir, 0755); err != nil {
		panic(fmt.Sprintf("Не удалось создать директорию для хранения device id: %v", err))
	}

	deviceidLocalCreatorOpts := deviceidcreator.LocalDeviceIdCreatorOptions{
		FileSys:          fileSystem,
		DeviceIdFilePath: domain.Path(deviceIdResolverConfig.GetDeviceIdFilePath()),
	}

	deviceidlocalCreator := deviceidcreator.NewLocalDeviceIdCreator(deviceidLocalCreatorOpts)
	localDeviceIdResolverOpts := deviceidlocal.LocalDeviceIdResolverOptions{
		LocalDeviceIdCreator: deviceidlocalCreator,
	}

	localIdDeviceResolver := deviceidlocal.NewLocalDeviceIdResolver(localDeviceIdResolverOpts)

	fileManagerOpts := filemanager.FileManagerOptions{
		RootResolver:          rootResolver,
		HashManager:           hashManager,
		FileSystem:            fileSystem,
		PathTreeWriter:        pathTree,
		LocalDeviceIdProvider: localIdDeviceResolver,

		Logger:    fileManagerLogger,
		LoggerCtx: appCtx,
	}

	fileManager := filemanager.NewFileManager(fileManagerOpts)

	fileUseCaseOpts := fileusecase.FileUseCaseOptions{
		FileManager: fileManager,
	}

	fileUseCase := fileusecase.NewFileUseCase(fileUseCaseOpts)

	serverLogger := logger.With(moduleAtrributeName, grpcServerModuleName)
	serverConfig := config.ServerConfig
	grpcServerOpts := server.GrpcServerOptions{
		FileUseCase: fileUseCase,

		CertPath:   serverConfig.TlsConfig.GetServerCertPath(),
		KeyPath:    serverConfig.TlsConfig.GetServerKeyPath(),
		CaCertPath: serverConfig.TlsConfig.GetCaCertPath(),

		NetworkType: serverConfig.NetworkType,
		Address:     serverConfig.GetAddress(),
		ServiceName: serverConfig.ServiceName,

		ChunkSizeInBytes: serverConfig.ChunkSizeInBytes,

		Ctx: appCtx,

		Logger: serverLogger,
	}

	grpcServer := server.NewGrpcServer(grpcServerOpts)
	err = grpcServer.Start()
	if err != nil {
		panic("Не удалось запустить gRPC сервер")
	}

	go func() {
		<-appCtx.Done()
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), grpcServerGracefulStopTimeoutSeconds*time.Second)
		defer timeoutCancel()

		err := grpcServer.Stop(timeoutCtx)
		if err != nil {
			logger.Warn("Не удалось коректно остановить grpcServer")
		}
	}()

	mDnsBrowserConfig := config.MDnsBrowserConfig
	mDnsBrowserLogger := logger.With(moduleAtrributeName, mDnsBrowserModuleName)

	mDnsNodeNamesBrowser, err := createMDnsNodeNamesBrowser(
		mDnsBrowserConfig.ServerServiceType,
		mDnsBrowserConfig.ServerDomain,
		mDnsBrowserConfig.GetInterfaces(),
		mDnsBrowserLogger,
		appCtx,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать MDnsNodeNamesBrowser: %v", err))
	}

	clientConfig := config.ClientConfig
	clientLogger := logger.With(moduleAtrributeName, grpcClientModuleName)

	// В GrpcConf не передается ServerAddress, т.к. он добавляется в ConnectionManager при подключении к конкретному узлу
	// Данная конфигурация является основой для подключения к любому узлу
	grpcClientConf := clt.GrpcConf{
		CertPath: clientConfig.TlsConfig.GetClientCertPath(),
		KeyPath:  clientConfig.TlsConfig.GetClientKeyPath(),

		CaCertPath: clientConfig.TlsConfig.GetCaCertPath(),

		ServerNetworkType: clientConfig.ServerNetworkType,
		ServerAddress:     clientConfig.ServerAddress,
		ServerServiceName: clientConfig.ServerServiceName,

		ResolverScheme:     clientConfig.ResolverScheme,
		MDnsResolverIfaces: mDnsBrowserConfig.GetInterfaces(),

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
	}

	connectionManagerLogger := logger.With(moduleAtrributeName, connectionManagerModuleName)

	mdnsUrldResolverOpts := mdnsurl.MDnsUrlResolverOptions{
		MDnsServerServiceType: mDnsBrowserConfig.ServerServiceType,
		MDnsServerDomain:      mDnsBrowserConfig.ServerDomain,
	}

	mdnsUrlResolver := mdnsurl.NewMDnsUrlResolver(mdnsUrldResolverOpts)

	connectionManagerOpts := conn.ConnectionManagerOptions{
		MDnsUrlResolver:  mdnsUrlResolver,
		BaseGrpcConf:     &grpcClientConf,
		GrpcClientLogger: clientLogger,
		Logger:           connectionManagerLogger,
		Ctx:              appCtx,
	}

	connectionManager := conn.NewConnectionManager(connectionManagerOpts)

	connectUseCaseOpts := connusecase.ConnectUseCaseOptions{

		ConnectionManager: connectionManager,
	}

	nodeUseCaseOpts := nodeusecase.NodeUseCaseOptions{
		NodeNamesBrowser: mDnsNodeNamesBrowser,
	}

	nodeUseCase := nodeusecase.NewNodeUseCase(nodeUseCaseOpts)

	connectUseCase := connusecase.NewConnectUseCase(connectUseCaseOpts)

	baseSnapshotRepositoryOpts := baserepo.BaseSnapshotRepositoryOptions{
		Db: db,
	}
	baseSnapshotRepository := baserepo.NewBaseSnapshotRepository(baseSnapshotRepositoryOpts)

	baseSnapshotProviderOpts := base.BaseSnapshotProviderOptions{
		BaseSnapshotRepository: baseSnapshotRepository,
	}
	baseSnapshotProvider := base.NewBaseSnapshotProvider(baseSnapshotProviderOpts)

	remoteSnapshotProviderOpts := remote.RemoteSnapshotProviderOptions{
		ClientFabric: connectionManager,
	}
	remoteSnapshotProvider := remote.NewRemoteSnapshotProvider(remoteSnapshotProviderOpts)

	localSnapshotProviderOpts := local.LocalSnapshotProviderOptions{
		FileManager: fileManager,
	}
	localSnapshotProvider := local.NewLocalSnapshotProvider(localSnapshotProviderOpts)

	planner := planner.NewChangesPlannerWithTreeSkip()

	planCache := planres.NewSyncPlanCache()

	planResolverOpts := planres.PlanResolverOptions{
		BaseSnapshotProvider:   baseSnapshotProvider,
		RemoteSnapshotProvider: remoteSnapshotProvider,
		LocalSnapshotProvider:  localSnapshotProvider,

		ChangesPlanner: planner,
		SyncPlanCache:  planCache,
	}

	planResolver := planres.NewPlanResolver(planResolverOpts)

	scanUseCaseOpts := scanusecase.ScanUseCaseOptions{
		PlanResolver: planResolver,
	}

	scanUseCase := scanusecase.NewScanUseCase(scanUseCaseOpts)

	rootUseCaseOpts := rootusecase.RootUseCaseOptions{
		RootRegistrar: rootResolver,
	}

	rootUseCase := rootusecase.NewRootUseCase(rootUseCaseOpts)

	cliLogger := logger.With(moduleAtrributeName, cliModuleName)

	cliInstanceOpts := cli.CliOptions{
		SyncUseCase:    nil, // TODO: добавить SyncUseCase
		ScanUseCase:    scanUseCase,
		NodeUseCase:    nodeUseCase,
		ConnectUseCase: connectUseCase,
		RootUseCase:    rootUseCase,

		Logger: cliLogger,

		Ctx: appCtx,
	}

	cliInstance := cli.NewCli(cliInstanceOpts)
	go func() { cliInstance.Start() }()

	<-stopSignal
	cancel()

	exitCtx, exitCancel := context.WithTimeout(context.Background(), exitTimeoutSeconds*time.Second)
	defer exitCancel()

	<-exitCtx.Done()
}
