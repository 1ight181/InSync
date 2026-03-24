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
	"google.golang.org/grpc/credentials"

	ifaces "insync/internal/interfaces"
	reslvr "insync/internal/transport/grpc/client/resolver"
)

const (
	resolverSchemeSeparator = "://"
)

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

	loadBalancingPolicy string

	ShouldUseHealthCheck bool

	ctx context.Context

	chunkSizeInBytes int

	logger *slog.Logger

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

	LoadBalancingPolicy string

	ShouldUseHealthCheck bool

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

		opts.Ctx == nil ||

		opts.ChunkSizeInBytes <= 0 ||

		opts.Logger == nil {
		panic("Все поля GrpcClientOption должны быть заполнены")
	}
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

		loadBalancingPolicy: opts.LoadBalancingPolicy,

		ctx: opts.Ctx,

		chunkSizeInBytes: opts.ChunkSizeInBytes,

		logger: opts.Logger,
	}
}

// Start устанавливает защищенное TLS соединение с gRPC сервером и создает gRPC клиент для взаимодействия с сервером.
// Если клиент уже запущен, возвращается ошибка.
// Этот метод должен быть вызван до использования клиента для выполнения RPC вызовов.
func (gc *GrpcClient) Start() error {
	if gc.isStarted.Swap(true) {
		gc.logger.Warn("Попытка запустить клиент, который уже запущен")
		return ErrClientAlreadyStarted
	}

	gc.logger.Info("Запуск gRPC клиента...")

	ctx := gc.ctx
	if ctx.Err() != nil {
		gc.isStarted.Store(false)
		return ctx.Err()
	}

	tlsConfig, err := gc.createClientTlsConfig()
	if err != nil {
		gc.isStarted.Store(false)
		return err
	}

	transportCreds := credentials.NewTLS(tlsConfig)
	withCreds := grpc.WithTransportCredentials(transportCreds)

	baseDialer := &net.Dialer{}
	networkAwareDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return baseDialer.DialContext(ctx, gc.serverNetworkType, addr)
	}
	withContextDialer := grpc.WithContextDialer(networkAwareDialer)

	stringBuilder := &strings.Builder{}
	stringBuilder.WriteString("{\"loadBalancingPolicy\": \"")
	stringBuilder.WriteString(gc.loadBalancingPolicy)
	if gc.ShouldUseHealthCheck {
		stringBuilder.WriteString(", \"healthCheckConfig\": { \"serviceName\":")
		stringBuilder.WriteString(gc.serverServiceName)
		stringBuilder.WriteString("\"}")
	}
	stringBuilder.WriteString("}")

	serviceConfig := stringBuilder.String()
	withDefaultServiceConf := grpc.WithDefaultServiceConfig(
		serviceConfig,
	)

	address := fmt.Sprintf("%s%s%s", gc.resolverScheme, resolverSchemeSeparator, gc.serverAddress)

	if ctx.Err() != nil {
		gc.isStarted.Store(false)
		return ctx.Err()
	}

	builderOptions := reslvr.BuilderOptions{
		ResolverIfaces:              gc.mdnsResolverIfaces,
		BackgroundListenTimeout:     time.Second * 30,
		ShouldResolveIpv6:           true,
		ShouldDisableResolverOnIdle: true,
		ShouldReportError:           true,
	}

	resolver := reslvr.NewBuilder(
		builderOptions,
	)

	withResolvers := grpc.WithResolvers(resolver)

	conn, err := grpc.NewClient(
		address,
		withCreds,
		withContextDialer,
		withDefaultServiceConf,
		withResolvers,
	)
	if err != nil {
		gc.isStarted.Store(false)
		return err
	}

	gc.clientConn = conn
	gc.client = insyncpb.NewFileSyncServiceClient(conn)

	gc.logger.Info("gRPC клиент успешно запущен")

	return nil
}

// Stop закрывает gRPC соединение с сервером. Если клиент не запущен, возвращается ошибка. Этот метод должен быть вызван для корректного завершения работы клиента и освобождения ресурсов.
func (gc *GrpcClient) Stop() error {
	if !gc.isStarted.Swap(false) {
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

	var fileList []domain.FileMetadata
	for _, file := range getFileListResponse.Files {
		fileMetadata := domain.FileMetadata{
			FileUuid:     file.FileUuid,
			RelativePath: file.RelativePath,
			IsDirectory:  file.IsDirectory,
			SizeBytes:    file.SizeBytes,
			ModifiedUnix: file.ModifiedUnix,
		}

		fileList = append(fileList, fileMetadata)
	}

	return fileList, nil
}

func (gc *GrpcClient) GetFile(ctx context.Context, rootName, relativePath string) (*io.PipeReader, error) {
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
				if err := pipeWriter.CloseWithError(err); errors.Is(err, io.ErrClosedPipe) {
					gc.logger.Warn(
						"Попытка повторно закрыть пайп методом CloseWithError при выполнении GetFile, который уже был закрыт",
					)
				}
				return
			}

			_, err = pipeWriter.Write(getFileResponse.Chunk.Data)
			if err != nil {
				if err := pipeWriter.CloseWithError(err); errors.Is(err, io.ErrClosedPipe) {
					gc.logger.Warn(
						"Попытка повторно закрыть пайп методом CloseWithError при выполнении GetFile, который уже был закрыт",
					)
				}
				return
			}
		}

		if err := pipeWriter.Close(); err != nil {
			gc.logger.LogAttrs(
				gc.ctx,
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
					gc.ctx,
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
	if !putFileResponse.Success || putFileResponse.Message != "" {
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
	if !deleteFileResponse.Success && deleteFileResponse.Message != "" {
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
		FileUuid:        fileUuid,
		NewRelativePath: relativePath,
	}

	renameFileResponse, err := gc.client.RenameFile(ctx, renameFileRequest)
	if err != nil {
		return err
	}
	if !renameFileResponse.Success && renameFileResponse.Message != "" {
		return RenameFileFailedError{
			Message: renameFileResponse.Message,
		}
	}

	return nil
}

func (gc *GrpcClient) createClientTlsConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(gc.certPath, gc.keyPath)
	if err != nil {
		return nil, err
	}

	gc.logger.LogAttrs(
		gc.ctx,
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
		gc.ctx,
		slog.LevelDebug,
		"CA сертификат успешно загружен для клиента",
		slog.String("caCertPath", gc.caCertPath),
	)

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
		ServerName:         gc.serverName,
	}, nil
}
