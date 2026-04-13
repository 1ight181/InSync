package app

import (
	"context"
	"fmt"
	clt "insync/internal/infrastructure/client"
	conn "insync/internal/infrastructure/connection"
	"insync/internal/infrastructure/filemanager"
	"insync/internal/infrastructure/filemanager/filesys"
	"insync/internal/infrastructure/filemanager/hash"
	"insync/internal/infrastructure/filemanager/pathtree"
	"insync/internal/infrastructure/filemanager/root"
	node "insync/internal/infrastructure/nodename"
	cli "insync/internal/presentation/cli"
	server "insync/internal/transport/grpc/server"
	connusecase "insync/internal/usecase/connect"
	fileusecase "insync/internal/usecase/file"
	syncusecase "insync/internal/usecase/sync"
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

	rootResolverOpts := root.NewRootResolver()
	fileSystemOpts := filesys.NewFileSystem()
	pathTree := pathtree.NewPathTree()
	hashCache := hash.NewHashCache()
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

	fileManagerConfig := config.FileManagerConfig
	fileManagerLogger := logger.With(moduleAtrributeName, fileManagerModuleName)

	fileManagerOpts := filemanager.FileManagerOptions{
		RootResolver:   rootResolverOpts,
		HashManager:    hashManager,
		FileSystem:     fileSystemOpts,
		PathTreeWriter: pathTree,

		TempDir:   fileManagerConfig.TempDir,
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
		CertPath:    serverConfig.TlsConfig.GetServerCertPath(),
		KeyPath:     serverConfig.TlsConfig.GetServerKeyPath(),
		CaCertPath:  serverConfig.TlsConfig.GetCaCertPath(),

		NetworkType: serverConfig.NetworkType,
		Address:     serverConfig.Address,
		ServiceName: serverConfig.ServiceName,

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

	nodeNameResolverOpts := node.NodeNameResolverOptions{
		MDnsServerServiceType: mDnsBrowserConfig.ServerServiceType,
		MDnsServerDomain:      mDnsBrowserConfig.ServerDomain,
	}

	nodeNameresolver := node.NewNodeNameResolver(nodeNameResolverOpts)

	connectionManagerOpts := conn.ConnectionManagerOptions{
		NodeNameResolver: nodeNameresolver,
		BaseGrpcConf:     &grpcClientConf,
		GrpcClientLogger: clientLogger,
		Logger:           connectionManagerLogger,
		Ctx:              appCtx,
	}

	connectionManager := conn.NewConnectionManager(connectionManagerOpts)

	connectUseCaseOpts := connusecase.ConnectUseCaseOptions{
		NodeNamesBrowser:  mDnsNodeNamesBrowser,
		ConnectionManager: connectionManager,
	}

	connectUseCase := connusecase.NewConnectUseCase(connectUseCaseOpts)

	syncUseCaseOpts := syncusecase.SyncUseCaseOptions{
		ClientFabric: connectionManager,
	}

	syncUseCase := syncusecase.NewSyncUseCase(syncUseCaseOpts)

	cliLogger := logger.With(moduleAtrributeName, cliModuleName)

	cliInstanceOpts := cli.CliOptions{
		SyncUseCase:    syncUseCase,
		ConnectUseCase: connectUseCase,

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
