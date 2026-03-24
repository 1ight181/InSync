package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	exitTimeoutSeconds   = 5
	moduleAtrributeName  = "module"
	mDnsModuleName       = "mdns"
	grpcServerModuleName = "grpc_server"
	grpcClientModuleName = "grpc_client"
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

	mDnsLogger := logger.With(moduleAtrributeName, mDnsModuleName)
	mDnsConfig := config.MDnsConfig

	if err := startMDnsServer(
		mDnsConfig.InstanceName,
		mDnsConfig.ServiceType,
		mDnsConfig.Domain,
		mDnsConfig.Port,
		mDnsConfig.GetInterfaces(),
		mDnsLogger,
		appCtx,
	); err != nil {
		panic(fmt.Sprintf("Не удалось запустить mDNS сервер: %v", err))
	}

	tlsConfig := config.TlsConfig

	serverLogger := logger.With(moduleAtrributeName, grpcServerModuleName)
	serverConfig := config.ServerConfig
	startGrpcServer(
		tlsConfig.GetServerCertPath(),
		tlsConfig.GetServerKeyPath(),
		tlsConfig.GetCaCertPath(),

		serverConfig.NetworkType,
		serverConfig.GetAddress(),

		appCtx,
		serverLogger,
	)

	clientConfig := config.ClientConfig

	clientLogger := logger.With(moduleAtrributeName, grpcClientModuleName)
	err = startGrpcClient(
		tlsConfig.GetClientCertPath(),
		tlsConfig.GetClientKeyPath(),
		tlsConfig.GetCaCertPath(),

		clientConfig.ServerNetworkType,
		clientConfig.ServerAddress,
		clientConfig.ServerServiceName,
		mDnsConfig.InstanceName,
		clientConfig.ResolverScheme,

		mDnsConfig.GetInterfaces(),

		clientConfig.LoadBalancingPolicy,

		appCtx,

		clientConfig.ChunkSizeInBytes,

		clientLogger,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось запустить gRPC клиент: %v", err))
	}

	// передавать в cli
	_, err = createMDnsBrowser(
		mDnsConfig.ServiceType,
		mDnsConfig.Domain,
		mDnsConfig.GetInterfaces(),
		mDnsLogger,
		appCtx,
	)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать mDNS браузер: %v", err))
	}

	<-stopSignal
	cancel()

	exitCtx, exitCancel := context.WithTimeout(context.Background(), exitTimeoutSeconds*time.Second)
	defer exitCancel()

	<-exitCtx.Done()
}
