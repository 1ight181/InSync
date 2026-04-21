package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
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

	resolver "google.golang.org/grpc/resolver"
)

const (
	defaultBaseDelay         = 100 * time.Millisecond
	defaultMultiplier        = 2.0
	defaultMaxDelay          = 10 * time.Second
	defaultJitter            = 0.1
	defaultMinConnectTimeout = 5 * time.Second
)

const (
	resolverSchemeSeparator = "://"
	defaultScheme           = "passthrough"
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

	LoadBalancingPolicy  string
	ShouldUseHealthCheck bool

	RpcTimeout       time.Duration
	RetryPolicy      *RpcRetryPolicy
	ConnectionConfig *ConnectionConfig

	ChunkSizeInBytes int

	DefaultServiceConf string
	Creds              credentials.TransportCredentials
	Dialer             func(context.Context, string) (net.Conn, error)
	Resolvers          []resolver.Builder
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

var (
	ErrInvalidOpts = errors.New("Все поля GrpcClientOption и GrpcConf в том числе должны быть заполнены")
)

func NewGrpcClient(opts GrpcClientOptions) (*GrpcClient, error) {
	if opts.Conf == nil {
		return nil, ErrInvalidOpts
	}

	if opts.Conf.Creds == nil && (opts.Conf.CaCertPath == "" ||
		opts.Conf.CertPath == "" ||
		opts.Conf.ServerName == "" ||
		opts.Conf.KeyPath == "") {
		return nil, ErrInvalidOpts
	}

	if opts.Conf.Dialer == nil &&
		opts.Conf.ServerNetworkType == "" {
		return nil, ErrInvalidOpts
	}

	if opts.Conf.ServerAddress == "" {
		return nil, ErrInvalidOpts
	}

	if len(opts.Conf.Resolvers) > 0 && opts.Conf.ResolverScheme == "" {
		return nil, ErrInvalidOpts
	}

	if opts.Conf.DefaultServiceConf == "" && (opts.Conf.LoadBalancingPolicy == "" ||
		opts.Conf.ServerServiceName == "" ||
		opts.Conf.RpcTimeout <= 0 ||
		opts.Conf.RetryPolicy == nil ||
		opts.Conf.RetryPolicy.MaxAttempts <= 0 ||
		opts.Conf.RetryPolicy.InitialBackoff <= 0 ||
		opts.Conf.RetryPolicy.MaxBackoff <= 0 ||
		opts.Conf.RetryPolicy.BackoffMultiplier <= 0 ||
		len(opts.Conf.RetryPolicy.RetryableStatusCodes) == 0) {
		return nil, ErrInvalidOpts
	}

	if opts.Ctx == nil ||

		opts.Conf.ChunkSizeInBytes <= 0 ||

		opts.Logger == nil {
		return nil, ErrInvalidOpts
	}

	loggerCtx := context.Background()
	return &GrpcClient{
		conf: opts.Conf,

		ctx: opts.Ctx,

		logger:    opts.Logger,
		loggerCtx: loggerCtx,
	}, nil
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

	var dialOptions []grpc.DialOption

	if gc.conf.Creds != nil {
		withTransportCreds := grpc.WithTransportCredentials(gc.conf.Creds)
		dialOptions = append(dialOptions, withTransportCreds)
	} else {
		transportCreds, err := gc.createTransportCreds()
		if err != nil {
			return err
		}

		withTransportCreds := grpc.WithTransportCredentials(transportCreds)
		dialOptions = append(dialOptions, withTransportCreds)
	}

	if gc.conf.Dialer != nil {
		dialOptions = append(dialOptions,
			grpc.WithContextDialer(gc.conf.Dialer),
		)
	} else {
		dialOptions = append(dialOptions,
			grpc.WithContextDialer(gc.createNetworkAwareDialer()),
		)
	}

	if len(gc.conf.Resolvers) > 0 {
		dialOptions = append(dialOptions,
			grpc.WithResolvers(gc.conf.Resolvers...),
		)
	}

	if gc.conf.DefaultServiceConf != "" {
		withDefaultServiceConf := grpc.WithDefaultServiceConfig(gc.conf.DefaultServiceConf)
		dialOptions = append(dialOptions, withDefaultServiceConf)
	} else {
		withDefaultServiceConf := grpc.WithDefaultServiceConfig(gc.createServiceConfig())
		dialOptions = append(dialOptions, withDefaultServiceConf)
	}

	if gc.conf.ConnectionConfig != nil {
		withConnectParams := grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  gc.conf.ConnectionConfig.BaseDelay,
				Multiplier: gc.conf.ConnectionConfig.Multiplier,
				MaxDelay:   gc.conf.ConnectionConfig.MaxDelay,
				Jitter:     gc.conf.ConnectionConfig.Jitter,
			},
			MinConnectTimeout: gc.conf.ConnectionConfig.MinConnectTimeout,
		})

		dialOptions = append(dialOptions, withConnectParams)
	} else {
		defaultConnectionConfig := ConnectionConfig{
			BaseDelay:         defaultBaseDelay,
			Multiplier:        defaultMultiplier,
			MaxDelay:          defaultMaxDelay,
			Jitter:            defaultJitter,
			MinConnectTimeout: defaultMinConnectTimeout,
		}
		withConnectParams := grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  defaultConnectionConfig.BaseDelay,
				Multiplier: defaultConnectionConfig.Multiplier,
				MaxDelay:   defaultConnectionConfig.MaxDelay,
				Jitter:     defaultConnectionConfig.Jitter,
			},
			MinConnectTimeout: defaultConnectionConfig.MinConnectTimeout,
		})
		dialOptions = append(dialOptions, withConnectParams)
	}

	ctx := gc.ctx
	if ctx.Err() != nil {
		return ctx.Err()
	}

	conn, err := grpc.NewClient(
		address,
		dialOptions...,
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
	scheme := gc.conf.ResolverScheme
	if scheme != "" && len(gc.conf.Resolvers) > 0 {
		return fmt.Sprintf("%s%s%s", scheme, resolverSchemeSeparator, gc.conf.ServerAddress)
	}

	return fmt.Sprintf("%s%s%s", defaultScheme, resolverSchemeSeparator, gc.conf.ServerAddress)
}
