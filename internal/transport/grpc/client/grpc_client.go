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

	ifaces "insync/internal/interfaces"
	mdnsresolver "insync/internal/transport/grpc/client/resolver"

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

type GrpcClient struct {
	certPath string
	keyPath  string

	caCertPath string

	serverNetworkType string
	serverAddress     string
	serverServiceName string
	serverName        string

	resolverScheme string

	// опциональный список интерфейсов, которые будут использоваться для работы mdns клиента
	mdnsResolverIfaces []string

	loadBalancingPolicy  string
	shouldUseHealthCheck bool

	rpcTimeout       time.Duration
	retryPolicy      *RpcRetryPolicy
	connectionConfig *ConnectionConfig

	ctx context.Context

	chunkSizeInBytes int

	logger    *slog.Logger
	loggerCtx context.Context

	clientConn *grpc.ClientConn

	client insyncpb.FileSyncServiceClient

	isStarted atomic.Bool
}

type GrpcClientOptions struct {
	CertPath   string
	KeyPath    string
	CaCertPath string

	NetworkType   string
	ServerAddress string
	ServiceName   string
	ServerName    string

	ResolverScheme string

	MdnsResolverIfaces []string

	LoadBalancingPolicy  string
	ShouldUseHealthCheck bool

	RpcTimeout       time.Duration
	RpcRetryPolicy   *RpcRetryPolicy
	ConnectionConfig *ConnectionConfig

	Ctx context.Context

	ChunkSizeInBytes int

	Logger *slog.Logger
}

func NewGrpcClient(opts GrpcClientOptions) ifaces.IClient {
	if opts.CertPath == "" ||
		opts.KeyPath == "" ||
		opts.CaCertPath == "" ||

		opts.NetworkType == "" ||
		opts.ServerAddress == "" ||
		opts.ServiceName == "" ||
		opts.ServerName == "" ||

		opts.ResolverScheme == "" ||

		opts.LoadBalancingPolicy == "" ||

		opts.RpcRetryPolicy == nil ||
		opts.ConnectionConfig == nil ||

		opts.Ctx == nil ||

		opts.ChunkSizeInBytes <= 0 ||

		opts.Logger == nil {
		panic("Все поля GrpcClientOption должны быть заполнены")
	}
	loggerCtx := context.Background()
	return &GrpcClient{
		certPath:   opts.CertPath,
		keyPath:    opts.KeyPath,
		caCertPath: opts.CaCertPath,

		serverNetworkType: opts.NetworkType,
		serverAddress:     opts.ServerAddress,
		serverServiceName: opts.ServiceName,
		serverName:        opts.ServerName,

		resolverScheme: opts.ResolverScheme,

		mdnsResolverIfaces: opts.MdnsResolverIfaces,

		loadBalancingPolicy:  opts.LoadBalancingPolicy,
		shouldUseHealthCheck: opts.ShouldUseHealthCheck,

		rpcTimeout:       opts.RpcTimeout,
		retryPolicy:      opts.RpcRetryPolicy,
		connectionConfig: opts.ConnectionConfig,

		ctx: opts.Ctx,

		chunkSizeInBytes: opts.ChunkSizeInBytes,

		logger:    opts.Logger,
		loggerCtx: loggerCtx,
	}
}

// Start устанавливает защищенное TLS соединение с gRPC сервером и создает gRPC клиент для взаимодействия с сервером.
// Если клиент уже запущен, возвращается ошибка.
// Этот метод должен быть вызван до использования клиента для выполнения RPC вызовов.
func (gc *GrpcClient) Start() (err error) {
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

	mdnsResolver := gc.createMdnsResolver()
	withResolvers := grpc.WithResolvers(mdnsResolver)

	withConnectParams := grpc.WithConnectParams(grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  gc.connectionConfig.BaseDelay,
			Multiplier: gc.connectionConfig.Multiplier,
			MaxDelay:   gc.connectionConfig.MaxDelay,
			Jitter:     gc.connectionConfig.Jitter,
		},
		MinConnectTimeout: gc.connectionConfig.MinConnectTimeout,
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

// Stop закрывает gRPC соединение с сервером. Если клиент не запущен, возвращается ошибка. Этот метод должен быть вызван для корректного завершения работы клиента и освобождения ресурсов.
func (gc *GrpcClient) Stop() error {
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
	cert, err := tls.LoadX509KeyPair(gc.certPath, gc.keyPath)
	if err != nil {
		return nil, err
	}

	gc.logger.LogAttrs(
		gc.loggerCtx,
		slog.LevelDebug,
		"Сертификат клиента успешно загружен",
		slog.String("certPath", gc.certPath),
		slog.String("keyPath", gc.keyPath),
	)

	caCert, err := os.ReadFile(gc.caCertPath)
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
		slog.String("caCertPath", gc.caCertPath),
	)

	return credentials.NewTLS(&tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
		ServerName:         gc.serverName,
	}), nil
}

func (gc *GrpcClient) createServiceConfig() string {

	retryableCodes := strings.Join(gc.retryPolicy.RetryableStatusCodes, `","`)
	retryableCodes = fmt.Sprintf(`["%s"]`, retryableCodes)

	healthCheckPart := ""
	if gc.shouldUseHealthCheck {
		healthCheckPart = fmt.Sprintf(`, "healthCheckConfig": { "serviceName": "%s" }`, gc.serverServiceName)
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
		gc.loadBalancingPolicy,
		healthCheckPart,
		gc.serverServiceName,
		gc.rpcTimeout,
		gc.retryPolicy.MaxAttempts,
		gc.retryPolicy.InitialBackoff,
		gc.retryPolicy.MaxBackoff,
		gc.retryPolicy.BackoffMultiplier,
		retryableCodes,
	)
}

func (gc *GrpcClient) createNetworkAwareDialer() func(context.Context, string) (net.Conn, error) {
	baseDialer := &net.Dialer{}
	networkAwareDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return baseDialer.DialContext(ctx, gc.serverNetworkType, addr)
	}
	return networkAwareDialer
}

func (gc *GrpcClient) createAddress() string {
	return fmt.Sprintf("%s%s%s", gc.resolverScheme, resolverSchemeSeparator, gc.serverAddress)
}

func (gc *GrpcClient) createMdnsResolver() resolver.Builder {
	builderOptions := mdnsresolver.BuilderOptions{
		ResolverIfaces:              gc.mdnsResolverIfaces,
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
