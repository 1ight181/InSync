package client

import (
	"context"
	"fmt"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
	"io"
	"net"
	"runtime/debug"
	"sync/atomic"

	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"

	ifaces "insync/internal/interfaces"
)

const (
	resolverSchemeSeparator = ":///"
)

type grpcClient struct {
	certPath string
	keyPath  string

	caCertPath string

	networkType   string
	serverAddress string

	resolverScheme string

	serviceName string

	ctx context.Context

	chunkSizeInBytes int

	logger *slog.Logger

	clientConn *grpc.ClientConn

	client insyncpb.FileSyncServiceClient

	isStarted  atomic.Bool
	isServerOk atomic.Bool
}

type grpcClientOptions struct {
	CertPath   string
	KeyPath    string
	CaCertPath string

	NetworkType    string
	ServerAddress  string
	ResolverScheme string

	ServiceName string

	Ctx context.Context

	ChunkSizeInBytes int

	Logger *slog.Logger
}

func newGrpcClient(opts grpcClientOptions) ifaces.IClient {
	if opts.CertPath == "" ||
		opts.KeyPath == "" ||
		opts.CaCertPath == "" ||

		opts.NetworkType == "" ||
		opts.ServerAddress == "" ||
		opts.ResolverScheme == "" ||

		opts.ServiceName == "" ||

		opts.Ctx == nil ||

		opts.ChunkSizeInBytes <= 0 ||

		opts.Logger == nil {
		panic("Все поля GrpcClientOption должны быть заполнены")
	}
	return &grpcClient{
		certPath:   opts.CertPath,
		keyPath:    opts.KeyPath,
		caCertPath: opts.CaCertPath,

		networkType:    opts.NetworkType,
		serverAddress:  opts.ServerAddress,
		resolverScheme: opts.ResolverScheme,

		serviceName: opts.ServiceName,

		ctx: opts.Ctx,

		chunkSizeInBytes: opts.ChunkSizeInBytes,

		logger: opts.Logger,
	}
}

func (gc *grpcClient) startHealthChecker(conn *grpc.ClientConn) error {
	healthCheckerClient := grpc_health_v1.NewHealthClient(conn)

	healthCheckStream, err := healthCheckerClient.Watch(gc.ctx, &grpc_health_v1.HealthCheckRequest{
		Service: gc.serviceName,
	})
	if err != nil {
		return err
	}

	go func() {
		var lastStatus grpc_health_v1.HealthCheckResponse_ServingStatus
		for {
			healthCheckResponse, err := healthCheckStream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					gc.isServerOk.Store(false)
					gc.logger.Info("Сервер завершил стрим своего состояния, вероятно сервер остановлен")
					break
				}
			}

			newStatus := healthCheckResponse.GetStatus()

			if lastStatus != newStatus {
				if newStatus == grpc_health_v1.HealthCheckResponse_SERVING {
					gc.isServerOk.Store(true)
				} else {
					gc.isServerOk.Store(false)
				}
			}

			gc.logger.LogAttrs(
				gc.ctx,
				slog.LevelDebug,
				"Состояние сервера изменилось",
				slog.String("last_status", lastStatus.String()),
				slog.String("new_status", newStatus.String()),
			)

		}
	}()

	return nil

}

func (gc *grpcClient) createClientTlsConfig() (*tls.Config, error) {
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
	}, nil
}

// Start устанавливает защищенное TLS соединение с gRPC сервером и создает gRPC клиент для взаимодействия с сервером.
// Если клиент уже запущен, возвращается ошибка.
// Этот метод должен быть вызван до использования клиента для выполнения RPC вызовов.
func (gc *grpcClient) Start() error {
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
	clientOptsWithCreds := grpc.WithTransportCredentials(transportCreds)

	baseDialer := &net.Dialer{}

	networkAwareDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return baseDialer.DialContext(ctx, gc.networkType, addr)
	}

	clientOptsWithDialer := grpc.WithContextDialer(networkAwareDialer)

	if ctx.Err() != nil {
		gc.isStarted.Store(false)
		return ctx.Err()
	}
	conn, err := grpc.NewClient(
		fmt.Sprintf(
			"%s%s%s",
			gc.resolverScheme,
			resolverSchemeSeparator,
			gc.serverAddress,
		),
		clientOptsWithCreds,
		clientOptsWithDialer,
	)
	if err != nil {
		gc.isStarted.Store(false)
		return err
	}

	gc.clientConn = conn
	gc.client = insyncpb.NewFileSyncServiceClient(conn)

	if err := gc.startHealthChecker(conn); err != nil {
		gc.isStarted.Store(false)
		return FailedToStartHealthCheckerError{Err: err}
	}

	gc.logger.Info("gRPC клиент успешно запущен")

	return nil
}

// Stop закрывает gRPC соединение с сервером. Если клиент не запущен, возвращается ошибка. Этот метод должен быть вызван для корректного завершения работы клиента и освобождения ресурсов.
func (gc *grpcClient) Stop() error {
	if !gc.isStarted.Swap(false) {
		gc.logger.Warn("Попытка остановить клиент, который не запущен")
		return ErrClientAlreadyStopped
	}

	if err := gc.clientConn.Close(); err != nil {
		return err
	}

	return nil
}

func (gc *grpcClient) GetFileList(ctx context.Context, rootName string) ([]domain.FileMetadata, error) {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда клиент не запущен")
		return nil, ErrClientNotStarted
	}

	if !gc.isServerOk.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда сервер недоступен")
		return nil, ServerUnavailableError{
			MethodName: "GetFileList",
		}
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

func (gc *grpcClient) GetFile(ctx context.Context, rootName, relativePath string) (*io.PipeReader, error) {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка получить файл, когда клиент не запущен")
		return nil, ErrClientNotStarted
	}

	if !gc.isServerOk.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда сервер недоступен")
		return nil, ServerUnavailableError{
			MethodName: "GetFile",
		}
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

func (gc *grpcClient) PutFile(ctx context.Context, file io.Reader, rootName, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка отправить файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	if !gc.isServerOk.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда сервер недоступен")
		return ServerUnavailableError{
			MethodName: "PutFile",
		}
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

func (gc *grpcClient) DeleteFile(ctx context.Context, rootName, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка удалить файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	if !gc.isServerOk.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда сервер недоступен")
		return ServerUnavailableError{
			MethodName: "DeleteFile",
		}
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

func (gc *grpcClient) RenameFile(ctx context.Context, rootName, fileUuid, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка переименовать файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	if !gc.isServerOk.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда сервер недоступен")
		return ServerUnavailableError{
			MethodName: "RenameFile",
		}
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
