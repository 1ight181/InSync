package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
	"io"
	"net"
	"os"
	"runtime/debug"
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

func (gc *GrpcClient) GetFileList(ctx context.Context, rootName string) ([]domain.FileMetadata, error) {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда клиент не запущен")
		return nil, ErrClientNotStarted
	}

	getFileListResponse, err := gc.client.GetFileList(ctx, &insyncpb.GetFileListRequest{
		RootName: rootName,
	})
	if err != nil {
		return nil, err
	}

	fileList := make([]domain.FileMetadata, 0, len(getFileListResponse.Files))
	for _, file := range getFileListResponse.Files {
		fileMetadata := domain.FileMetadata{
			RelativePath: file.RelativePath,
			IsDirectory:  file.IsDirectory,
			SizeBytes:    file.SizeBytes,
			ModifiedUnix: file.ModifiedUnix,
		}

		fileList = append(fileList, fileMetadata)
	}

	return fileList, nil
}

func (gc *GrpcClient) GetFile(ctx context.Context, rootName, relativePath string) (io.ReadCloser, error) {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка получить файл, когда клиент не запущен")
		return nil, ErrClientNotStarted
	}

	getFileRequest := &insyncpb.GetFileRequest{
		RootName:     rootName,
		RelativePath: relativePath,
	}

	getFileResponseStream, err := gc.client.GetFile(ctx, getFileRequest)
	if err != nil {
		return nil, err
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		for {
			getFileResponse, err := getFileResponseStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				if closeErr := pipeWriter.CloseWithError(err); errors.Is(closeErr, io.ErrClosedPipe) {
					gc.logger.Warn(
						"Попытка повторно закрыть пайп методом CloseWithError при выполнении GetFile, который уже был закрыт",
					)
				}
				return
			}

			writeCompleted := make(chan struct{})

			go func() {
				defer close(writeCompleted)
				_, writeErr := pipeWriter.Write(getFileResponse.Chunk.Data)
				if writeErr != nil {
					if closeErr := pipeWriter.CloseWithError(writeErr); errors.Is(closeErr, io.ErrClosedPipe) {
						gc.logger.Warn(
							"Попытка повторно закрыть пайп методом CloseWithError при выполнении GetFile, который уже был закрыт",
						)
					}
					return
				}
			}()

			select {
			case <-ctx.Done():
				pipeWriter.CloseWithError(ctx.Err())
				if err := getFileResponseStream.CloseSend(); err != nil {
					gc.logger.LogAttrs(
						gc.loggerCtx,
						slog.LevelWarn,
						"Ошибка при закрытии стрима методом CloseSend при выполнении GetFile",
						slog.String("error", err.Error()),
						slog.String("trace", string(debug.Stack())),
					)
				}
				return
			case <-writeCompleted:
			}
		}

		if err := pipeWriter.Close(); err != nil {
			gc.logger.LogAttrs(
				gc.loggerCtx,
				slog.LevelWarn,
				"Ошибка при закрытии пайпа методом Close при выполнении GetFile",
				slog.String("error", err.Error()),
				slog.String("trace", string(debug.Stack())),
			)
		}
	}()

	return pipeReader, nil
}

func (gc *GrpcClient) PutFile(ctx context.Context, file io.Reader, rootName, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка отправить файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	putFileStream, err := gc.client.PutFile(ctx)
	if err != nil {
		return err
	}

	initMessage := &insyncpb.PutFileRequest{
		Payload: &insyncpb.PutFileRequest_Init{
			Init: &insyncpb.PutFileInit{
				RootName:     rootName,
				RelativePath: relativePath,
			},
		},
	}

	err = putFileStream.Send(initMessage)
	if err != nil {
		return err
	}

	buf := make([]byte, gc.chunkSizeInBytes)

	for i := 0; ; i++ {
		numberOfBytes, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			if closeErr := putFileStream.CloseSend(); closeErr != nil {
				gc.logger.LogAttrs(
					gc.loggerCtx,
					slog.LevelError,
					"Ошибка при закрытии потока PutFile после получения ошибки от чтения файла",
					slog.String("readError", err.Error()),
					slog.String("closeSendError", closeErr.Error()),
					slog.String("trace", string(debug.Stack())),
				)
			}
			return err
		}

		chunkMessage := &insyncpb.PutFileRequest{
			Payload: &insyncpb.PutFileRequest_Chunk{
				Chunk: &insyncpb.FileChunk{
					Index: uint32(i),
					Data:  buf[:numberOfBytes],
				},
			},
		}

		err = putFileStream.Send(chunkMessage)
		if err != nil {
			return err
		}
	}

	putFileResponse, err := putFileStream.CloseAndRecv()
	if err != nil {
		return err
	}
	if !putFileResponse.Success {
		return PutFileFailedError{
			Message: putFileResponse.Message,
		}
	}

	return nil
}

func (gc *GrpcClient) DeleteFile(ctx context.Context, rootName, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка удалить файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	deleteFileRequest := &insyncpb.DeleteFileRequest{
		RootName:     rootName,
		RelativePath: relativePath,
	}

	deleteFileResponse, err := gc.client.DeleteFile(ctx, deleteFileRequest)
	if err != nil {
		return err
	}
	if !deleteFileResponse.Success {
		return DeleteFileFailedError{
			Message: deleteFileResponse.Message,
		}
	}

	return nil
}

func (gc *GrpcClient) RenameFile(ctx context.Context, rootName, fileUuid, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка переименовать файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	renameFileRequest := &insyncpb.RenameFileRequest{
		RootName:        rootName,
		NewRelativePath: relativePath,
	}

	renameFileResponse, err := gc.client.RenameFile(ctx, renameFileRequest)
	if err != nil {
		return err
	}
	if !renameFileResponse.Success {
		return RenameFileFailedError{
			Message: renameFileResponse.Message,
		}
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
