package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"insync/internal/syncproto"
	"log/slog"
	"net"
	"os"

	serverr "insync/internal/server/errors"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcServer struct {
	syncproto.UnimplementedFileSyncServiceServer
	serverCertPath string
	serverKeyPath  string
	caCertPath     string

	networkType string
	address     string

	ctx context.Context

	logger slog.Logger

	server *grpc.Server
}

type GrpcServerOption struct {
	Ctx            context.Context
	Logger         slog.Logger
	ServerCertPath string
	ServerKeyPath  string
	CaCertPath     string
	NetworkType    string
	Address        string
}

func NewGrpcServer(opts GrpcServerOption) *GrpcServer {
	if opts.Ctx == nil ||
		opts.Logger == (slog.Logger{}) ||
		opts.ServerCertPath == "" ||
		opts.ServerKeyPath == "" ||
		opts.CaCertPath == "" ||
		opts.NetworkType == "" ||
		opts.Address == "" {
		panic("Все поля GrpcServerOption должны быть заполнены")
	}
	return &GrpcServer{
		ctx:            opts.Ctx,
		logger:         opts.Logger,
		serverCertPath: opts.ServerCertPath,
		serverKeyPath:  opts.ServerKeyPath,
		caCertPath:     opts.CaCertPath,
		networkType:    opts.NetworkType,
		address:        opts.Address,
	}
}

func (s *GrpcServer) createServerTlsConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(s.serverCertPath, s.serverKeyPath)
	if err != nil {
		return nil, err
	}

	s.logger.LogAttrs(
		s.ctx,
		slog.LevelDebug,
		"Сертификат сервера успешно загружен",
		slog.String("certPath", s.serverCertPath),
		slog.String("keyPath", s.serverKeyPath),
	)

	caCert, err := os.ReadFile(s.caCertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	isOk := caCertPool.AppendCertsFromPEM(caCert)
	if !isOk {
		return nil, serverr.ErrFailedToAppendCa
	}

	s.logger.LogAttrs(
		s.ctx,
		slog.LevelDebug,
		"CA сертификат успешно загружен",
		slog.String("caCertPath", s.caCertPath),
	)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}

// Start блокирует выполнение, поэтому его нужно запускать в отдельной горутине. Он будет работать до тех пор, пока сервер не будет остановлен через метод Stop.
func (s *GrpcServer) Start() error {
	s.logger.Info("Запуск gRPC сервера...")
	tlsConfig, err := s.createServerTlsConfig()
	if err != nil {
		return serverr.ServerStartError{Err: err}
	}

	listener, err := net.Listen(s.networkType, s.address)
	if err != nil {
		return serverr.ServerStartError{Err: err}
	}

	creds := credentials.NewTLS(tlsConfig)

	server := grpc.NewServer(grpc.Creds(creds))
	s.server = server

	syncproto.RegisterFileSyncServiceServer(server, s)
	if err := server.Serve(listener); err != nil && err != grpc.ErrServerStopped {
		return serverr.ServerStartError{Err: err}
	}

	return nil
}

// Stop инициирует процесс остановки сервера. Он сначала пытается выполнить GracefulStop,
// который позволяет завершить текущие соединения и запросы.
// Если в течение заданного таймаута сервер не успевает остановиться,
// вызывается принудительная остановка через Stop.
func (s *GrpcServer) Stop(timeoutCtx context.Context) {
	s.logger.Info("Остановка gRPC сервера...")
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("gRPC сервер успешно остановлен")
	case <-timeoutCtx.Done():
		s.server.Stop()
		s.logger.Warn("Вызвана принудительная остановка gRPC сервера из-за таймаута")
	}
}
