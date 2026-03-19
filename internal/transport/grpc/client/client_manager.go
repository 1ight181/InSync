package client

import (
	"context"
	"errors"
	ifaces "insync/internal/interfaces"
	"log/slog"
	"net"
	"syscall"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcClientManager struct {
	certPath string
	keyPath  string

	caCertPath string

	networkType     string
	serverAddresses []string

	resolverScheme string

	serviceName string

	ctx context.Context

	chunkSizeInBytes int

	logger *slog.Logger

	client ifaces.IClient
}

type GrpcClientManagerOptions struct {
	CertPath         string
	KeyPath          string
	CaCertPath       string
	NetworkType      string
	ServerAddresses  []string
	ResolverScheme   string
	ServiceName      string
	Ctx              context.Context
	ChunkSizeInBytes int
	Logger           *slog.Logger
}

func NewGrpcClientManager(opts GrpcClientManagerOptions) *GrpcClientManager {
	if opts.CertPath == "" ||
		opts.KeyPath == "" ||
		opts.CaCertPath == "" ||
		opts.NetworkType == "" ||
		opts.ServerAddresses == nil ||
		opts.ResolverScheme == "" ||
		opts.ServiceName == "" ||
		opts.Ctx == nil ||
		opts.ChunkSizeInBytes <= 0 ||
		opts.Logger == nil {
		panic("Все поля GrpcClientManagerOptions должны быть заполнены")
	}
	return &GrpcClientManager{
		certPath:         opts.CertPath,
		keyPath:          opts.KeyPath,
		caCertPath:       opts.CaCertPath,
		networkType:      opts.NetworkType,
		serverAddresses:  opts.ServerAddresses,
		resolverScheme:   opts.ResolverScheme,
		serviceName:      opts.ServiceName,
		ctx:              opts.Ctx,
		chunkSizeInBytes: opts.ChunkSizeInBytes,
		logger:           opts.Logger,
		client:           nil,
	}
}

func (cm *GrpcClientManager) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Проверка gRPC ошибок
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unavailable,
			codes.DeadlineExceeded:
			return true
		default:
			return false
		}
	}

	// Проверка TCP / сетевых ошибок
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}

	// Проверка системных ошибок
	if errors.Is(err, syscall.ECONNREFUSED) || // connection refused
		errors.Is(err, syscall.ENETUNREACH) || // network unreachable
		errors.Is(err, syscall.EHOSTUNREACH) { // no route to host
		return true
	}

	return false
}

func (cm *GrpcClientManager) GetClient() (ifaces.IClient, error) {
	addresses := cm.serverAddresses
	var client ifaces.IClient
	for _, address := range addresses {
		clientOpts := grpcClientOptions{
			CertPath:   cm.certPath,
			KeyPath:    cm.keyPath,
			CaCertPath: cm.caCertPath,

			NetworkType: cm.networkType,

			ServerAddress:    address,
			ResolverScheme:   cm.resolverScheme,
			ServiceName:      cm.serviceName,
			Ctx:              cm.ctx,
			ChunkSizeInBytes: cm.chunkSizeInBytes,
			Logger:           cm.logger,
		}

		client = newGrpcClient(clientOpts)

		if err := client.Start(); cm.isRetryable(err) {
			cm.logger.Warn("Повторная попытка подключения к gRPC серверу", "error", err.Error())
			continue
		} else if err != nil {
			return nil, err
		}
		cm.client = client
		return cm.client, nil
	}

	return nil, FailedToConnectToServerError{
		Addresses: addresses,
	}
}
