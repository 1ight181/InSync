package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"insync/internal/transport/grpc/insyncpb"
	"log/slog"
	"net"
	"os"
	"sync/atomic"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GrpcServer struct {
	insyncpb.UnimplementedFileSyncServiceServer

	fileUseCase IFileUseCase
	baseUseCase IBaseSnapshotUseCase

	creds credentials.TransportCredentials

	certPath   string
	keyPath    string
	caCertPath string

	networkType string
	address     string
	serviceName string

	chunkSizeInBytes int

	logger    *slog.Logger
	loggerCtx context.Context

	listener net.Listener

	server *grpc.Server

	shouldStartHealthServer bool
	healthServer            *health.Server

	isStarted atomic.Bool
}

type GrpcServerOptions struct {
	FileUseCase IFileUseCase
	BaseUseCase IBaseSnapshotUseCase

	Creds credentials.TransportCredentials

	CertPath   string
	KeyPath    string
	CaCertPath string

	NetworkType string
	Address     string
	ServiceName string

	ShouldStartHealthServer bool
	Listener                net.Listener

	ChunkSizeInBytes int
	Logger           *slog.Logger
}

var (
	ErrInvalidGrpcServerOptions = errors.New("Все поля GrpcServerOptions должны быть заполнены")
)

func NewGrpcServer(opts GrpcServerOptions) (*GrpcServer, error) {
	if opts.Creds == nil && (opts.CertPath == "" ||
		opts.KeyPath == "" ||
		opts.CaCertPath == "") {
		return nil, ErrInvalidGrpcServerOptions
	}

	if opts.FileUseCase == nil ||
		opts.BaseUseCase == nil ||
		opts.NetworkType == "" ||
		opts.Address == "" ||
		opts.ServiceName == "" ||

		opts.ChunkSizeInBytes <= 0 ||

		opts.Logger == nil {
		return nil, ErrInvalidGrpcServerOptions
	}
	loggerCtx := context.Background()
	return &GrpcServer{
		fileUseCase: opts.FileUseCase,
		baseUseCase: opts.BaseUseCase,

		certPath:   opts.CertPath,
		keyPath:    opts.KeyPath,
		caCertPath: opts.CaCertPath,

		networkType: opts.NetworkType,
		address:     opts.Address,
		serviceName: opts.ServiceName,

		chunkSizeInBytes: opts.ChunkSizeInBytes,

		creds:    opts.Creds,
		listener: opts.Listener,

		logger:    opts.Logger,
		loggerCtx: loggerCtx,
	}, nil
}

func (gs *GrpcServer) Start() (err error) {
	if !gs.isStarted.CompareAndSwap(false, true) {
		gs.logger.Warn("Попытка запустить сервер, который уже запущен")
		return ErrServerAlreadyStarted
	}
	defer func() {
		if err != nil {
			gs.isStarted.Store(false)
		}
	}()

	gs.logger.Info("Запуск gRPC сервера...")

	var withTransportCreds grpc.ServerOption
	if gs.creds != nil {
		withTransportCreds = grpc.Creds(gs.creds)
	} else {
		transportCreds, err := gs.createTransportCreds()
		if err != nil {
			return err
		}
		withTransportCreds = grpc.Creds(transportCreds)
	}

	server := grpc.NewServer(withTransportCreds)
	gs.server = server
	insyncpb.RegisterFileSyncServiceServer(server, gs)

	var healthServer *health.Server
	if gs.shouldStartHealthServer {
		healthServer = health.NewServer()
		gs.healthServer = healthServer
		grpc_health_v1.RegisterHealthServer(server, healthServer)
	}

	var listener net.Listener
	if gs.listener != nil {
		listener = gs.listener
	} else {
		listener, err = net.Listen(gs.networkType, gs.address)
		if err != nil {
			return err
		}
	}

	go func() {
		defer gs.isStarted.Store(false)
		if err := server.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			gs.logger.LogAttrs(
				gs.loggerCtx,
				slog.LevelError,
				"Произошла ошибка при запуске gRPC сервера",
				slog.String("err", err.Error()),
			)
		}
	}()

	if gs.shouldStartHealthServer {
		healthServer.SetServingStatus(gs.serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	}
	gs.logger.Info("gRPC сервер успешно запущен")

	return nil
}

// Stop инициирует процесс остановки сервера. Он сначала пытается выполнить GracefulStop,
// который позволяет завершить текущие соединения и запросы.
// Если в течение заданного таймаута сервер не успевает остановиться,
// вызывается принудительная остановка через Stop.
func (gs *GrpcServer) Stop(timeoutCtx context.Context) error {
	if !gs.isStarted.Swap(false) {
		gs.logger.Warn("Попытка остановить сервер, который не запущен")
		return ErrServerAlreadyStopped
	}
	gs.logger.Info("Остановка gRPC сервера...")
	done := make(chan struct{})
	go func() {
		gs.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		gs.logger.Info("gRPC сервер успешно остановлен")
	case <-timeoutCtx.Done():
		gs.server.Stop()
		gs.logger.Warn("Вызвана принудительная остановка gRPC сервера из-за таймаута")
	}

	return nil
}

func (gs *GrpcServer) createTransportCreds() (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(gs.certPath, gs.keyPath)
	if err != nil {
		return nil, err
	}

	gs.logger.LogAttrs(
		gs.loggerCtx,
		slog.LevelDebug,
		"Сертификат сервера успешно загружен",
		slog.String("certPath", gs.certPath),
		slog.String("keyPath", gs.keyPath),
	)

	caCert, err := os.ReadFile(gs.caCertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	isOk := caCertPool.AppendCertsFromPEM(caCert)
	if !isOk {
		return nil, ErrFailedToAppendCa
	}

	gs.logger.LogAttrs(
		gs.loggerCtx,
		slog.LevelDebug,
		"CA сертификат успешно загружен для сервера",
		slog.String("caCertPath", gs.caCertPath),
	)

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}), nil
}
