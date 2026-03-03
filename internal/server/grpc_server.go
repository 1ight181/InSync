package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"insync/internal/insyncpb"
	"log/slog"
	"net"
	"os"
	"sync/atomic"

	serverr "insync/internal/server/errors"
	serverifaces "insync/internal/server/interfaces"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GrpcServer struct {
	insyncpb.UnimplementedFileSyncServiceServer
	serverCertPath string
	serverKeyPath  string
	caCertPath     string

	serverNetworkType string
	serverAddress     string

	ctx context.Context

	logger *slog.Logger

	server *grpc.Server

	isStarted atomic.Bool
}

type GrpcServerOptions struct {
	ServerCertPath string
	ServerKeyPath  string
	CaCertPath     string

	ServerNetworkType string
	ServerAddress     string

	Ctx context.Context

	Logger *slog.Logger
}

func NewGrpcServer(opts GrpcServerOptions) serverifaces.IServer {
	if opts.ServerCertPath == "" ||
		opts.ServerKeyPath == "" ||
		opts.CaCertPath == "" ||

		opts.ServerNetworkType == "" ||
		opts.ServerAddress == "" ||

		opts.Ctx == nil ||

		opts.Logger == nil {
		panic("Все поля GrpcServerOption должны быть заполнены")
	}
	return &GrpcServer{
		serverCertPath: opts.ServerCertPath,
		serverKeyPath:  opts.ServerKeyPath,
		caCertPath:     opts.CaCertPath,

		serverNetworkType: opts.ServerNetworkType,
		serverAddress:     opts.ServerAddress,

		ctx: opts.Ctx,

		logger: opts.Logger,
	}
}

func (gs *GrpcServer) createServerTlsConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(gs.serverCertPath, gs.serverKeyPath)
	if err != nil {
		return nil, err
	}

	gs.logger.LogAttrs(
		gs.ctx,
		slog.LevelDebug,
		"Сертификат сервера успешно загружен",
		slog.String("certPath", gs.serverCertPath),
		slog.String("keyPath", gs.serverKeyPath),
	)

	caCert, err := os.ReadFile(gs.caCertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	isOk := caCertPool.AppendCertsFromPEM(caCert)
	if !isOk {
		return nil, serverr.ErrFailedToAppendCa
	}

	gs.logger.LogAttrs(
		gs.ctx,
		slog.LevelDebug,
		"CA сертификат успешно загружен для сервера",
		slog.String("caCertPath", gs.caCertPath),
	)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}

// Start блокирует выполнение, поэтому его нужно запускать в отдельной горутине. Он будет работать до тех пор, пока сервер не будет остановлен через метод Stop.
func (gs *GrpcServer) Start() error {
	if gs.isStarted.Swap(true) {
		gs.logger.Warn("Попытка запустить сервер, который уже запущен")
		return serverr.ErrServerAlreadyStarted
	}
	defer gs.isStarted.Store(false)

	ctx := gs.ctx
	if ctx.Err() != nil {
		return serverr.ServerStartError{Err: ctx.Err()}
	}

	gs.logger.Info("Запуск gRPC сервера...")
	tlsConfig, err := gs.createServerTlsConfig()
	if err != nil {
		return serverr.ServerStartError{Err: err}
	}

	listener, err := net.Listen(gs.serverNetworkType, gs.serverAddress)
	if err != nil {
		return serverr.ServerStartError{Err: err}
	}

	transportCreds := credentials.NewTLS(tlsConfig)
	serverOptsWithCreds := grpc.Creds(transportCreds)

	server := grpc.NewServer(serverOptsWithCreds)
	gs.server = server

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	insyncpb.RegisterFileSyncServiceServer(server, gs)

	if ctx.Err() != nil {
		return serverr.ServerStartError{Err: ctx.Err()}
	}
	if err := server.Serve(listener); err != nil && err != grpc.ErrServerStopped {
		return serverr.ServerStartError{Err: err}
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
		return serverr.ErrServerAlreadyStopped
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
