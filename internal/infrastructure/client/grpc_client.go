package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"insync/internal/transport/grpc/insyncpb"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"

	mdnsresolver "insync/internal/infrastructure/client/resolver"

	resolver "google.golang.org/grpc/resolver"
)

const (
	resolverSchemeSeparator = "://"
)

type RpcRetryPolicy struct {
	MaxAttempts          int
	InitialBackoff       time.Duration
	MaxBackoff           time.Duration
	BackoffMultiplier    float64
	RetryableStatusCodes []string
}

type ConnectionConfig struct {
	BaseDelay         time.Duration
	Multiplier        float64
	MaxDelay          time.Duration
	Jitter            float64
	MinConnectTimeout time.Duration
}

type GrpcConf struct {
	CertPath string
	KeyPath  string

	CaCertPath string

	ServerNetworkType string
	ServerAddress     string
	ServerServiceName string
	ServerName        string

	ResolverScheme string

	// опциональный список интерфейсов, которые будут использоваться для работы mdns клиента
	MDnsResolverIfaces []string

	LoadBalancingPolicy  string
	ShouldUseHealthCheck bool

	RpcTimeout       time.Duration
	RetryPolicy      *RpcRetryPolicy
	ConnectionConfig *ConnectionConfig

	ChunkSizeInBytes int
}

type GrpcClient struct {
	conf *GrpcConf

	ctx context.Context

	logger    *slog.Logger
	loggerCtx context.Context

	clientConn *grpc.ClientConn

	client insyncpb.FileSyncServiceClient

	isStarted atomic.Bool
}

type GrpcClientOptions struct {
	Conf *GrpcConf

	Ctx context.Context

	Logger *slog.Logger
}

func NewGrpcClient(opts GrpcClientOptions) *GrpcClient {
	if opts.Conf == nil ||
		opts.Conf.CertPath == "" ||
		opts.Conf.KeyPath == "" ||
		opts.Conf.CaCertPath == "" ||

		opts.Conf.ServerNetworkType == "" ||
		opts.Conf.ServerAddress == "" ||
		opts.Conf.ServerServiceName == "" ||
		opts.Conf.ServerName == "" ||

		opts.Conf.ResolverScheme == "" ||

		opts.Conf.LoadBalancingPolicy == "" ||

		opts.Conf.RetryPolicy == nil ||
		opts.Conf.ConnectionConfig == nil ||

		opts.Ctx == nil ||

		opts.Conf.ChunkSizeInBytes <= 0 ||

		opts.Logger == nil {
		panic("Все поля GrpcClientOption и GrpcConf в том числе должны быть заполнены")
	}
	loggerCtx := context.Background()
	return &GrpcClient{
		conf: opts.Conf,

		ctx: opts.Ctx,

		logger:    opts.Logger,
		loggerCtx: loggerCtx,
	}
}

// Connect устанавливает защищенное TLS соединение с gRPC сервером и создает gRPC клиент для взаимодействия с сервером.
// Если клиент уже запущен, возвращается ошибка.
// Этот метод должен быть вызван до использования клиента для выполнения RPC вызовов.
func (gc *GrpcClient) Connect() (err error) {
	defer func() {
		if err != nil {
			gc.isStarted.Store(false)
		}
	}()
	if !gc.isStarted.CompareAndSwap(false, true) {
		gc.logger.Warn("Попытка запустить клиент, который уже запущен")
		return ErrClientAlreadyStarted
	}

	gc.logger.Info("Запуск gRPC клиента...")

	address := gc.createAddress()

	gc.logger.LogAttrs(gc.loggerCtx, slog.LevelDebug, "Клиент подключается по адресу:", slog.String("address", address))

	transportCreds, err := gc.createTransportCreds()
	if err != nil {
		return err
	}
	withTransportCreds := grpc.WithTransportCredentials(transportCreds)

	networkAwareDialer := gc.createNetworkAwareDialer()
	withContextDialer := grpc.WithContextDialer(networkAwareDialer)

	serviceConfig := gc.createServiceConfig()
	withDefaultServiceConf := grpc.WithDefaultServiceConfig(
		serviceConfig,
	)

	gc.logger.LogAttrs(gc.loggerCtx, slog.LevelDebug, "Клиент будет запущен с следующим serviceConfig:", slog.String("serviceConfig", serviceConfig))

	mdnsResolver := gc.createMDnsResolver()
	withResolvers := grpc.WithResolvers(mdnsResolver)

	withConnectParams := grpc.WithConnectParams(grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  gc.conf.ConnectionConfig.BaseDelay,
			Multiplier: gc.conf.ConnectionConfig.Multiplier,
			MaxDelay:   gc.conf.ConnectionConfig.MaxDelay,
			Jitter:     gc.conf.ConnectionConfig.Jitter,
		},
		MinConnectTimeout: gc.conf.ConnectionConfig.MinConnectTimeout,
	})

	ctx := gc.ctx
	if ctx.Err() != nil {
		return ctx.Err()
	}

	conn, err := grpc.NewClient(
		address,
		withTransportCreds,
		withContextDialer,
		withDefaultServiceConf,
		withResolvers,
		withConnectParams,
	)
	if err != nil {
		return err
	}

	gc.clientConn = conn
	gc.client = insyncpb.NewFileSyncServiceClient(conn)

	gc.logger.Info("gRPC клиент успешно запущен")

	return nil
}

// Close закрывает gRPC соединение с сервером. Если клиент не запущен, возвращается ошибка. Этот метод должен быть вызван для корректного завершения работы клиента и освобождения ресурсов.
func (gc *GrpcClient) Close() error {
	if !gc.isStarted.CompareAndSwap(true, false) {
		gc.logger.Warn("Попытка остановить клиент, который не запущен")
		return ErrClientAlreadyStopped
	}

	if err := gc.clientConn.Close(); err != nil {
		return err
	}

	return nil
}

func (gc *GrpcClient) createTransportCreds() (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(gc.conf.CertPath, gc.conf.KeyPath)
	if err != nil {
		return nil, err
	}

	gc.logger.LogAttrs(
		gc.loggerCtx,
		slog.LevelDebug,
		"Сертификат клиента успешно загружен",
		slog.String("certPath", gc.conf.CertPath),
		slog.String("keyPath", gc.conf.KeyPath),
	)

	caCert, err := os.ReadFile(gc.conf.CaCertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, ErrFailedToAppendCa
	}

	gc.logger.LogAttrs(
		gc.loggerCtx,
		slog.LevelDebug,
		"CA сертификат успешно загружен для клиента",
		slog.String("caCertPath", gc.conf.CaCertPath),
	)

	return credentials.NewTLS(&tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
		ServerName:         gc.conf.ServerName,
	}), nil
}

func (gc *GrpcClient) createServiceConfig() string {

	retryableCodes := strings.Join(gc.conf.RetryPolicy.RetryableStatusCodes, `","`)
	retryableCodes = fmt.Sprintf(`["%s"]`, retryableCodes)

	healthCheckPart := ""
	if gc.conf.ShouldUseHealthCheck {
		healthCheckPart = fmt.Sprintf(`, "healthCheckConfig": { "serviceName": "%s" }`, gc.conf.ServerServiceName)
	}

	return fmt.Sprintf(`{
        "loadBalancingPolicy": "%s"%s,
        "methodConfig": [{
            "name": [{"service": "%s"}],
			"timeout": "%s",
            "retryPolicy": {
                "MaxAttempts": %d,
                "InitialBackoff": "%s",
                "MaxBackoff": "%s",
                "BackoffMultiplier": %f,
                "RetryableStatusCodes": %s
            }

        }]
    }`,
		gc.conf.LoadBalancingPolicy,
		healthCheckPart,
		gc.conf.ServerServiceName,
		gc.conf.RpcTimeout,
		gc.conf.RetryPolicy.MaxAttempts,
		gc.conf.RetryPolicy.InitialBackoff,
		gc.conf.RetryPolicy.MaxBackoff,
		gc.conf.RetryPolicy.BackoffMultiplier,
		retryableCodes,
	)
}

func (gc *GrpcClient) createNetworkAwareDialer() func(context.Context, string) (net.Conn, error) {
	baseDialer := &net.Dialer{}
	networkAwareDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return baseDialer.DialContext(ctx, gc.conf.ServerNetworkType, addr)
	}
	return networkAwareDialer
}

func (gc *GrpcClient) createAddress() string {
	return fmt.Sprintf("%s%s%s", gc.conf.ResolverScheme, resolverSchemeSeparator, gc.conf.ServerAddress)
}

func (gc *GrpcClient) createMDnsResolver() resolver.Builder {
	builderOptions := mdnsresolver.BuilderOptions{
		ResolverIfaces:              gc.conf.MDnsResolverIfaces,
		BackgroundListenTimeout:     time.Second * 30,
		ShouldResolveIpv6:           true,
		ShouldDisableResolverOnIdle: true,
		ShouldReportError:           true,
		Logger:                      gc.logger,
	}

	return mdnsresolver.NewBuilder(
		builderOptions,
	)
}
