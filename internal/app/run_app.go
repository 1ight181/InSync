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
	exitTimeoutSeconds = 5
)

func RunApp() {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	appCtx, cancel := context.WithCancel(context.Background())

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

	serverLogger := logger.With("module", "grpc_server")

	tlsConfig := config.TlsConfig

	serverConfig := config.ServerConfig
	startGrpcServer(
		tlsConfig.GetServerCertPath(),
		tlsConfig.GetServerKeyPath(),
		tlsConfig.GetCaCertPath(),

		serverConfig.ServerNetworkType,
		serverConfig.GetServerAddress(),

		appCtx,
		serverLogger,
	)

	clientConfig := config.ClientConfig
	err = startGrpcClient(
		tlsConfig.GetClientCertPath(),
		tlsConfig.GetClientKeyPath(),
		tlsConfig.GetCaCertPath(),

		clientConfig.ClientNetworkType,
		clientConfig.GetClientAddress(),

		appCtx,

		clientConfig.ChunkSizeInBytes,

		logger,
	)
	if err != nil {
		panic("Не удалось запустить Grpc клиента")
	}

	<-stopSignal
	cancel()

	exitCtx, exitCancel := context.WithTimeout(context.Background(), exitTimeoutSeconds*time.Second)
	defer exitCancel()

	<-exitCtx.Done()
}
