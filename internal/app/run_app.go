package app

import (
	"context"
	"fmt"
	"insync/internal/domain"
	clt "insync/internal/infrastructure/client"
	mdnsresolver "insync/internal/infrastructure/client/resolver"
	conn "insync/internal/infrastructure/connection"
	deviceid "insync/internal/infrastructure/deviceid"
	deviceidlocal "insync/internal/infrastructure/deviceid/local"
	deviceidcreator "insync/internal/infrastructure/deviceid/local/creator"
	deviceidremote "insync/internal/infrastructure/deviceid/remote"
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
	syncer "insync/internal/infrastructure/syncer"
	changeappl "insync/internal/infrastructure/syncer/applier"
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
	syncusecase "insync/internal/usecase/sync"
	"os"
	"time"

	"google.golang.org/grpc/resolver"
)

const (
	grpcServerGracefulStopTimeoutSeconds = 4
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
	config, err := createConfig()
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

	db, err := repo.NewGorm(dbOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать gorm: %v", err))
	}

	defer func() {
		sqlDb, err := db.DB()
		if err != nil {
			logger.Warn("Не удалось получить sqlDb для закрытия sqlite")
		}
		if err := sqlDb.Close(); err != nil {
			logger.Warn("Не удалось коректно закрыть sqlite")
		}
	}()

	hashCacheRepoOpts := hashrepo.HashRepositoryOptions{
		Db: db,
	}

	hashCacheRepo, err := hashrepo.NewHashRepository(hashCacheRepoOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать HashRepository: %v", err))
	}

	hashCacheOpts := hashcache.HashCacheOptions{
		HashCacheRepository: hashCacheRepo,
	}

	hashCache, err := hashcache.NewHashCache(hashCacheOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать HashCache: %v", err))
	}
	hashCalc := hash.NewHashCalculator()

	hashManagerLogger := logger.With(moduleAtrributeName, hashManagerModuleName)

	hashManagerOpts := hash.HashManagerOptions{
		HashCache:      hashCache,
		HashCalculator: hashCalc,
		PathTreeReader: pathTree,
		Logger:         hashManagerLogger,
	}

	hashManager, err := hash.NewHashManager(hashManagerOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать HashManager: %v", err))
	}

	fileManagerLogger := logger.With(moduleAtrributeName, fileManagerModuleName)

	deviceIdResolverConfig := config.DeviceIdResolverConfig
	deviceIdDir := deviceIdResolverConfig.DeviceIdDir
	if err := os.MkdirAll(deviceIdDir, 0755); err != nil {
		panic(fmt.Sprintf("Не удалось создать директорию для хранения device id: %v", err))
	}

	fileManagerOpts := filemanager.FileManagerOptions{
		RootResolver:   rootResolver,
		HashManager:    hashManager,
		FileSystem:     fileSystem,
		PathTreeWriter: pathTree,

		Logger: fileManagerLogger,
	}

	fileManager, err := filemanager.NewFileManager(fileManagerOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать fileManager: %v", err))
	}

	fileUseCaseOpts := fileusecase.FileUseCaseOptions{
		FileManager: fileManager,
	}

	fileUseCase, err := fileusecase.NewFileUseCase(fileUseCaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать fileUseCase: %v", err))
	}

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

		Logger: serverLogger,
	}

	grpcServer, err := server.NewGrpcServer(grpcServerOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать gRPC сервер: %v", err))
	}
	err = grpcServer.Start()
	if err != nil {
		panic("Не удалось запустить gRPC сервер")
	}

	defer func() {
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
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать MDnsNodeNamesBrowser: %v", err))
	}

	clientConfig := config.ClientConfig
	clientLogger := logger.With(moduleAtrributeName, grpcClientModuleName)

	builderOptions := mdnsresolver.BuilderOptions{
		ResolverIfaces:              mDnsBrowserConfig.GetInterfaces(),
		BackgroundListenTimeout:     time.Second * 30,
		ShouldResolveIpv6:           true,
		ShouldDisableResolverOnIdle: true,
		ShouldReportError:           true,
		Logger:                      logger,
	}

	mDnsResolverBuilder := mdnsresolver.NewBuilder(
		builderOptions,
	)

	// В GrpcConf не передается ServerAddress, т.к. он добавляется в ConnectionManager при подключении к конкретному узлу
	// Данная конфигурация является основой для подключения к любому узлу
	grpcClientConf := clt.GrpcConf{
		CertPath: clientConfig.TlsConfig.GetClientCertPath(),
		KeyPath:  clientConfig.TlsConfig.GetClientKeyPath(),

		CaCertPath: clientConfig.TlsConfig.GetCaCertPath(),

		ServerNetworkType: clientConfig.ServerNetworkType,
		ServerServiceName: clientConfig.ServerServiceName,

		ResolverScheme: clientConfig.ResolverScheme,

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
		Resolvers:        []resolver.Builder{mDnsResolverBuilder},
	}

	connectionManagerLogger := logger.With(moduleAtrributeName, connectionManagerModuleName)

	mdnsUrldResolverOpts := mdnsurl.MDnsUrlResolverOptions{
		MDnsServerServiceType: mDnsBrowserConfig.ServerServiceType,
		MDnsServerDomain:      mDnsBrowserConfig.ServerDomain,
	}

	mdnsUrlResolver, err := mdnsurl.NewMDnsUrlResolver(mdnsUrldResolverOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать MDnsUrlResolver: %v", err))
	}

	connectionManagerOpts := conn.ConnectionManagerOptions{
		MDnsUrlResolver:  mdnsUrlResolver,
		BaseGrpcConf:     &grpcClientConf,
		GrpcClientLogger: clientLogger,
		Logger:           connectionManagerLogger,
	}
	connectionManager, err := conn.NewConnectionManager(connectionManagerOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать ConnectionManager: %v", err))
	}

	connectUseCaseOpts := connusecase.ConnectUseCaseOptions{

		ConnectionManager: connectionManager,
	}

	nodeUseCaseOpts := nodeusecase.NodeUseCaseOptions{
		NodeNamesBrowser: mDnsNodeNamesBrowser,
	}

	nodeUseCase, err := nodeusecase.NewNodeUseCase(nodeUseCaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать NodeUseCase: %v", err))
	}

	connectUseCase, err := connusecase.NewConnectUseCase(connectUseCaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать ConnectUseCase: %v", err))
	}

	baseSnapshotRepositoryOpts := baserepo.BaseSnapshotRepositoryOptions{
		Db: db,
	}
	baseSnapshotRepository, err := baserepo.NewBaseSnapshotRepository(baseSnapshotRepositoryOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать BaseSnapshotRepository: %v", err))
	}

	deviceidLocalCreatorOpts := deviceidcreator.LocalDeviceIdCreatorOptions{
		FileSys:          fileSystem,
		DeviceIdFilePath: domain.Path(deviceIdResolverConfig.GetDeviceIdFilePath()),
	}

	deviceidlocalCreator, err := deviceidcreator.NewLocalDeviceIdCreator(deviceidLocalCreatorOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать LocalDeviceIdCreator: %v", err))
	}
	localDeviceIdResolverOpts := deviceidlocal.LocalDeviceIdResolverOptions{
		LocalDeviceIdCreator: deviceidlocalCreator,
	}

	localIdDeviceResolver, err := deviceidlocal.NewLocalDeviceIdResolver(localDeviceIdResolverOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать LocalDeviceIdResolver: %v", err))
	}

	remoteDeviceIdResolverConfig := config.RemoteDeviceIdResolverConfig

	remoteDeviceIdResolverOpts := deviceidremote.RemoteDeviceIdResolverOptions{
		MDnsServerInstanceNamePostfix: remoteDeviceIdResolverConfig.MDnsServerInstanceNamePostfix,
		NodeNameProvider:              connectionManager,
	}
	remoteDeviceIdResolver, err := deviceidremote.NewRemoteDeviceIdResolver(remoteDeviceIdResolverOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать RemoteDeviceIdResolver: %v", err))
	}

	deviceIdProviderOpts := deviceid.DeviceIdProviderOptions{
		LocalDeviceIdResolver:  localIdDeviceResolver,
		RemoteDeviceIdResolver: remoteDeviceIdResolver,
	}

	deviceIdProvider, err := deviceid.NewDeviceIdProvider(deviceIdProviderOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать DeviceIdProvider: %v", err))
	}

	baseSnapshotProviderOpts := base.BaseSnapshotProviderOptions{
		BaseSnapshotRepository: baseSnapshotRepository,
		DeviceIdProvider:       deviceIdProvider,
	}
	baseSnapshotProvider, err := base.NewBaseSnapshotProvider(baseSnapshotProviderOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать BaseSnapshotProvider: %v", err))
	}

	remoteSnapshotProviderOpts := remote.RemoteSnapshotProviderOptions{
		ClientFactory: connectionManager,
	}
	remoteSnapshotProvider, err := remote.NewRemoteSnapshotProvider(remoteSnapshotProviderOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать RemoteSnapshotProvider: %v", err))
	}

	localSnapshotProviderOpts := local.LocalSnapshotProviderOptions{
		FileManager: fileManager,
	}
	localSnapshotProvider, err := local.NewLocalSnapshotProvider(localSnapshotProviderOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать LocalSnapshotProvider: %v", err))
	}

	planner := planner.NewChangesPlannerWithTreeSkip()

	planCache := planres.NewSyncPlanCache()

	planResolverOpts := planres.PlanResolverOptions{
		BaseSnapshotProvider:   baseSnapshotProvider,
		RemoteSnapshotProvider: remoteSnapshotProvider,
		LocalSnapshotProvider:  localSnapshotProvider,

		ChangesPlanner: planner,
		SyncPlanCache:  planCache,
	}

	planResolver, err := planres.NewPlanResolver(planResolverOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать PlanResolver: %v", err))
	}

	scanUseCaseOpts := scanusecase.ScanUseCaseOptions{
		PlanResolver: planResolver,
	}
	scanUseCase, err := scanusecase.NewScanUseCase(scanUseCaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать ScanUseCase: %v", err))
	}

	rootUseCaseOpts := rootusecase.RootUseCaseOptions{
		RootRegistrar: rootResolver,
	}
	rootUseCase, err := rootusecase.NewRootUseCase(rootUseCaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать RootUseCase: %v", err))
	}

	changeApplierOpts := changeappl.ChangeApplierOptions{
		FileManager:   fileManager,
		ClientFactory: connectionManager,
	}
	changeApplier, err := changeappl.NewChangeApplier(changeApplierOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать ChangeApplier: %v", err))
	}

	conflictResolver := syncer.NewConflictResolver()

	syncerOpts := syncer.SyncerOptions{
		ChangeApplier:    changeApplier,
		ConflictResolver: conflictResolver,
	}
	syncer, err := syncer.NewSyncer(syncerOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать Syncer: %v", err))
	}

	syncusecaseOpts := syncusecase.SyncUseCaseOptions{
		SyncPlanResolver: planResolver,
		Syncer:           syncer,
	}

	syncUseCase, err := syncusecase.NewSyncUseCase(syncusecaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать SyncUseCase: %v", err))
	}

	cliLogger := logger.With(moduleAtrributeName, cliModuleName)

	cliInstanceOpts := cli.CliOptions{
		SyncUseCase:    syncUseCase,
		ScanUseCase:    scanUseCase,
		NodeUseCase:    nodeUseCase,
		ConnectUseCase: connectUseCase,
		RootUseCase:    rootUseCase,

		Logger: cliLogger,
	}

	cliInstance, err := cli.NewCli(cliInstanceOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать Cli: %v", err))
	}
	cliInstance.Start()
}
