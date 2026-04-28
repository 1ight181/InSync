package app

import (
	"context"
	"fmt"
	"insync/internal/domain"
	base "insync/internal/infrastructure/base"
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
	"insync/internal/infrastructure/holder"
	local "insync/internal/infrastructure/local"
	planner "insync/internal/infrastructure/planner"
	planres "insync/internal/infrastructure/planresolver"
	remote "insync/internal/infrastructure/remote"
	"insync/internal/infrastructure/root"
	syncer "insync/internal/infrastructure/syncer"
	changeappl "insync/internal/infrastructure/syncer/applier"
	persister "insync/internal/infrastructure/syncer/persister"
	cli "insync/internal/presentation/cli"
	repo "insync/internal/repository/sqlite"
	baserepo "insync/internal/repository/sqlite/base"
	hashrepo "insync/internal/repository/sqlite/hash"
	rootrepo "insync/internal/repository/sqlite/root"
	server "insync/internal/transport/grpc/server"
	baseusecase "insync/internal/usecase/base"
	connusecase "insync/internal/usecase/connect"
	fileusecase "insync/internal/usecase/file"
	initusecase "insync/internal/usecase/init"
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
	dbModuleName                         = "db"
)

type cleanupStack []func()

func (stack *cleanupStack) Add(cleanup func()) {
	*stack = append(*stack, cleanup)
}

func (stack cleanupStack) Run() <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		for cleanupIndex := len(stack) - 1; cleanupIndex >= 0; cleanupIndex-- {
			stack[cleanupIndex]()
		}
		close(doneChan)
	}()

	return doneChan
}

func RunApp() {
	var cleanups cleanupStack

	config, err := createConfig()
	if err != nil {
		panic(fmt.Sprintf("Не удалось загрузить и проверить конфиг: %v", err))
	}

	loggerConfig := config.LoggerConfig
	logDir := loggerConfig.LogFileDirectory
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("Не удалось создать директорию для логов: %v", err))
	}
	logger, err := createLogger(
		loggerConfig.ShouldLogToFile,
		loggerConfig.GetLogFilePath(),
		loggerConfig.LogLevel,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать логгер: %v", err))
	}

	dbConfig := config.DbConfig
	dbDir := dbConfig.Dir
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		panic(fmt.Sprintf("Не удалось создать директорию для БД: %v", err))
	}
	dsn := dbConfig.GetDsn()

	dbLogger := logger.With(moduleAtrributeName, dbModuleName)

	migrator := repo.NewAutoMigrator()
	dbOpts := repo.SqliteGormOptions{
		Dsn:      dsn,
		Migrator: migrator,
		Logger:   dbLogger,
	}

	db, err := repo.NewGorm(dbOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать gorm: %v", err))
	}

	cleanups.Add(func() {
		sqlDb, err := db.DB()
		if err != nil {
			logger.Warn("Не удалось получить sqlDb для закрытия sqlite")
		}
		if err := sqlDb.Close(); err != nil {
			logger.Warn("Не удалось коректно закрыть sqlite")
		}

	})

	rootRepo := rootrepo.NewRootRepository(db)

	rootResolver, err := root.NewRootResolver(rootRepo)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать RootResolver: %v", err))
	}
	fileSystem := filesys.NewFileSystem()
	pathTree := pathtree.NewPathTree()

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
		HashCache:            hashCache,
		HashCalculator:       hashCalc,
		PathTreeReader:       pathTree,
		DirtyPathsRepository: hashCacheRepo,
		Logger:               hashManagerLogger,
	}

	hashManager, err := hash.NewHashManager(hashManagerOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать HashManager: %v", err))
	}

	fileManagerLogger := logger.With(moduleAtrributeName, fileManagerModuleName)

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

	deviceIdResolverConfig := config.DeviceIdResolverConfig
	deviceIdDir := deviceIdResolverConfig.DeviceIdDir
	if err := os.MkdirAll(deviceIdDir, 0755); err != nil {
		panic(fmt.Sprintf("Не удалось создать директорию для хранения device id: %v", err))
	}

	deviceIdFilePathRaw := deviceIdResolverConfig.GetDeviceIdFilePath()
	deviceIdFilePath, err := domain.NewPath(deviceIdFilePathRaw)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать путь для хранения device id: %v", err))
	}

	if err := os.MkdirAll(deviceIdResolverConfig.DeviceIdDir, 0755); err != nil {
		panic(fmt.Sprintf("Не удалось создать директорию для : %v", err))
	}

	deviceidLocalCreatorOpts := deviceidcreator.LocalDeviceIdCreatorOptions{
		FileSys:          fileSystem,
		DeviceIdFilePath: deviceIdFilePath,
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

	clientConfig := config.ClientConfig
	clientLogger := logger.With(moduleAtrributeName, grpcClientModuleName)

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

	builderOptions := mdnsresolver.BuilderOptions{
		ResolverIfaces:    mDnsBrowserConfig.GetInterfaces(),
		ShouldResolveIpv6: true,
		ShouldReportError: true,
		Logger:            logger,
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

	clientHolder := holder.NewClientHolder()

	connectionManagerOpts := conn.ConnectionManagerOptions{
		BaseGrpcConf:      &grpcClientConf,
		ClientHolder:      clientHolder,
		ServerServiceType: clientConfig.MdnsServerServiceType,
		ServerNamePrefix:  clientConfig.MdnsServerNamePrefix,
		GrpcClientLogger:  clientLogger,
		Logger:            connectionManagerLogger,
	}
	connectionManager, err := conn.NewConnectionManager(connectionManagerOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать ConnectionManager: %v", err))
	}

	remoteDeviceIdResolverOpts := deviceidremote.RemoteDeviceIdResolverOptions{
		NodeNameProvider: connectionManager,
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

	baseSnapshotRepositoryOpts := baserepo.BaseSnapshotRepositoryOptions{
		Db: db,
	}
	baseSnapshotRepository, err := baserepo.NewBaseSnapshotRepository(baseSnapshotRepositoryOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать BaseSnapshotRepository: %v", err))
	}

	baseSnapshotManagerOpts := base.BaseSnapshotManagerOptions{
		BaseSnapshotRepository: baseSnapshotRepository,
		DeviceIdProvider:       deviceIdProvider,
	}
	baseSnapshotManager, err := base.NewBaseSnapshotManager(baseSnapshotManagerOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать BaseSnapshotManager: %v", err))
	}

	localSnapshotProviderOpts := local.LocalSnapshotProviderOptions{
		FileManager: fileManager,
	}
	localSnapshotProvider, err := local.NewLocalSnapshotProvider(localSnapshotProviderOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать LocalSnapshotProvider: %v", err))
	}

	baseUseCaseOpts := baseusecase.BaseSnapshotUseCaseOpts{
		BaseSnapshotCreator:   baseSnapshotManager,
		LocalSnapshotProvider: localSnapshotProvider,
	}

	baseUseCase, err := baseusecase.NewBaseSnapshotUseCase(baseUseCaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать BaseSnapshotUseCase: %v", err))
	}

	serverLogger := logger.With(moduleAtrributeName, grpcServerModuleName)
	serverConfig := config.ServerConfig
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
		panic(fmt.Sprintf("Не удалось создать gRPC сервер: %v", err))
	}
	err = grpcServer.Start()
	if err != nil {
		panic("Не удалось запустить gRPC сервер")
	}

	cleanups.Add(func() {
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), grpcServerGracefulStopTimeoutSeconds*time.Second)
		defer timeoutCancel()

		err := grpcServer.Stop(timeoutCtx)
		if err != nil {
			logger.Warn("Не удалось коректно остановить grpcServer")
		}
	})

	connectUseCaseOpts := connusecase.ConnectUseCaseOptions{
		ConnectionManager: connectionManager,
	}

	connectUseCase, err := connusecase.NewConnectUseCase(connectUseCaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать ConnectUseCase: %v", err))
	}

	nodeUseCaseOpts := nodeusecase.NodeUseCaseOptions{
		NodeNamesBrowser:      mDnsNodeNamesBrowser,
		LocalDeviceIdResolver: localIdDeviceResolver,
	}

	nodeUseCase, err := nodeusecase.NewNodeUseCase(nodeUseCaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать NodeUseCase: %v", err))
	}

	localDeviceId, err := localIdDeviceResolver.Resolve()
	if err != nil {
		panic(fmt.Sprintf("Не удалось получить device id: %v", err))
	}

	mDnsServerInstanceName := localDeviceId.String()

	mDnsServerLogger := logger.With(moduleAtrributeName, mDnsServerModuleName)
	mDnsServerConfig := config.MDnsServerConfig

	if err := startMDnsServer(
		mDnsServerInstanceName,
		mDnsServerConfig.ServiceType,
		mDnsServerConfig.Domain,
		mDnsServerConfig.Port,
		mDnsServerConfig.GetInterfaces(),
		mDnsServerLogger,
	); err != nil {
		panic(fmt.Sprintf("Не удалось запустить mDNS сервер: %v", err))
	}

	remoteSnapshotProviderOpts := remote.RemoteSnapshotProviderOptions{
		ClientFactory: connectionManager,
	}
	remoteSnapshotProvider, err := remote.NewRemoteSnapshotProvider(remoteSnapshotProviderOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать RemoteSnapshotProvider: %v", err))
	}

	planner := planner.NewChangesPlannerWithTreeSkip()

	planCache := planres.NewSyncPlanCache()

	planResolverOpts := planres.PlanResolverOptions{
		BaseSnapshotProvider:   baseSnapshotManager,
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

	postSyncBaseSnapshotPersisterOpts := persister.PostSyncBaseSnapshotPersisterOptions{
		BaseSnapshotRepository: baseSnapshotRepository,
		DeviceIdProvider:       deviceIdProvider,
		LocalSnapshotProvider:  localSnapshotProvider,
		ClientFactory:          connectionManager,
	}
	postSyncBaseSnapshotPersister, err := persister.NewPostSyncBaseSnapshotPersister(postSyncBaseSnapshotPersisterOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать PostSyncBaseSnapshotPersister: %v", err))
	}

	syncerOpts := syncer.SyncerOptions{
		ChangeApplier:                 changeApplier,
		ConflictResolver:              conflictResolver,
		PostSyncBaseSnapshotPersister: postSyncBaseSnapshotPersister,
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

	initUseCaseOpts := initusecase.InitUseCaseOptions{
		BaseSnapshotCreator:    baseSnapshotManager,
		RemoteSnapshotProvider: remoteSnapshotProvider,
		LocalSnapshotProvider:  localSnapshotProvider,
	}
	initUseCase, err := initusecase.NewInitUseCase(initUseCaseOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать InitUseCase: %v", err))
	}

	cliInstanceOpts := cli.CliOptions{
		SyncUseCase:    syncUseCase,
		ScanUseCase:    scanUseCase,
		NodeUseCase:    nodeUseCase,
		ConnectUseCase: connectUseCase,
		RootUseCase:    rootUseCase,
		InitUseCase:    initUseCase,

		Logger: cliLogger,
	}

	cliInstance, err := cli.NewCli(cliInstanceOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать Cli: %v", err))
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
