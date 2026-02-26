package client

import (
	"context"
	"insync/internal/models"
	syncproto "insync/internal/syncproto"
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

	clienterr "insync/internal/client/errors"
	clientifaces "insync/internal/client/interfaces"
)

type GrpcClient struct {
	clientCertPath string
	clientKeyPath  string
	caCertPath     string

	clientNetworkType string
	clientAddress     string

	ctx context.Context

	chunkSizeInBytes int

	logger *slog.Logger

	clientConn *grpc.ClientConn

	client syncproto.FileSyncServiceClient

	isStarted atomic.Bool
}

type GrpcClientOptions struct {
	ClientCertPath string
	ClientKeyPath  string
	CaCertPath     string

	ClientNetworkType string
	ClientAddress     string

	Ctx context.Context

	ChunkSizeInBytes int

	Logger *slog.Logger
}

func NewGrpcClient(opts GrpcClientOptions) clientifaces.IClient {
	if opts.ClientCertPath == "" ||
		opts.ClientKeyPath == "" ||
		opts.CaCertPath == "" ||

		opts.ClientNetworkType == "" ||
		opts.ClientAddress == "" ||

		opts.Ctx == nil ||

		opts.ChunkSizeInBytes <= 0 ||

		opts.Logger == nil {
		panic("Все поля GrpcClientOption должны быть заполнены")
	}
	return &GrpcClient{
		clientCertPath: opts.ClientCertPath,
		clientKeyPath:  opts.ClientKeyPath,
		caCertPath:     opts.CaCertPath,

		clientNetworkType: opts.ClientNetworkType,
		clientAddress:     opts.ClientAddress,

		ctx: opts.Ctx,

		chunkSizeInBytes: opts.ChunkSizeInBytes,

		logger: opts.Logger,
	}
}

func (gc *GrpcClient) createClientTlsConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(gc.clientCertPath, gc.clientKeyPath)
	if err != nil {
		return nil, err
	}

	gc.logger.LogAttrs(
		gc.ctx,
		slog.LevelDebug,
		"Сертификат клиента успешно загружен",
		slog.String("certPath", gc.clientCertPath),
		slog.String("keyPath", gc.clientKeyPath),
	)

	caCert, err := os.ReadFile(gc.caCertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, clienterr.ErrFailedToAppendCa
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
func (gc *GrpcClient) Start() error {
	if gc.isStarted.Swap(true) {
		gc.logger.Warn("Попытка запустить клиент, который уже запущен")
		return clienterr.ErrClientAlreadyStarted
	}

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

	networkAwareDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		baseDialer := &net.Dialer{}

		return baseDialer.DialContext(ctx, gc.clientNetworkType, addr)
	}

	clientOptsWithDialer := grpc.WithContextDialer(networkAwareDialer)

	if ctx.Err() != nil {
		gc.isStarted.Store(false)
		return ctx.Err()
	}
	conn, err := grpc.NewClient(gc.clientAddress, clientOptsWithCreds, clientOptsWithDialer)
	if err != nil {
		gc.isStarted.Store(false)
		return err
	}

	gc.clientConn = conn
	gc.client = syncproto.NewFileSyncServiceClient(conn)

	return nil
}

// Stop закрывает gRPC соединение с сервером. Если клиент не запущен, возвращается ошибка. Этот метод должен быть вызван для корректного завершения работы клиента и освобождения ресурсов.
func (gc *GrpcClient) Stop() error {
	if !gc.isStarted.Swap(false) {
		gc.logger.Warn("Попытка остановить клиент, который не запущен")
		return clienterr.ErrClientAlreadyStopped
	}

	if err := gc.clientConn.Close(); err != nil {
		return err
	}

	return nil
}

func (gc *GrpcClient) GetFileList(ctx context.Context, rootName string) ([]models.FileMetadata, error) {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда клиент не запущен")
		return nil, clienterr.ErrClientNotStarted
	}

	getFileListResponse, err := gc.client.GetFileList(ctx, &syncproto.GetFileListRequest{
		RootName: rootName,
	})
	if err != nil {
		return nil, err
	}

	var fileList []models.FileMetadata
	for _, file := range getFileListResponse.Files {
		fileMetadata := models.FileMetadata{
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
		return nil, clienterr.ErrClientNotStarted
	}

	getFileRequest := &syncproto.GetFileRequest{
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
		return clienterr.ErrClientNotStarted
	}

	putFileStream, err := gc.client.PutFile(ctx)
	if err != nil {
		return err
	}

	initMessage := &syncproto.PutFileRequest{
		Payload: &syncproto.PutFileRequest_Init{
			Init: &syncproto.PutFileInit{
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

		chunkMessage := &syncproto.PutFileRequest{
			Payload: &syncproto.PutFileRequest_Chunk{
				Chunk: &syncproto.FileChunk{
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
		return clienterr.PutFileFailedError{
			Message: putFileResponse.Message,
		}
	}

	return nil
}

func (gc *GrpcClient) DeleteFile(ctx context.Context, rootName, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка удалить файл, когда клиент не запущен")
		return clienterr.ErrClientNotStarted
	}

	deleteFileRequest := &syncproto.DeleteFileRequest{
		RootName:     rootName,
		RelativePath: relativePath,
	}

	deleteFileResponse, err := gc.client.DeleteFile(ctx, deleteFileRequest)
	if err != nil {
		return err
	}
	if !deleteFileResponse.Success && deleteFileResponse.Message != "" {
		return clienterr.DeleteFileFailedError{
			Message: deleteFileResponse.Message,
		}
	}

	return nil
}

func (gc *GrpcClient) RenameFile(ctx context.Context, rootName, fileUuid, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка переименовать файл, когда клиент не запущен")
		return clienterr.ErrClientNotStarted
	}

	renameFileRequest := &syncproto.RenameFileRequest{
		RootName:        rootName,
		FileUuid:        fileUuid,
		NewRelativePath: relativePath,
	}

	renameFileResponse, err := gc.client.RenameFile(ctx, renameFileRequest)
	if err != nil {
		return err
	}
	if !renameFileResponse.Success && renameFileResponse.Message != "" {
		return clienterr.RenameFileFailedError{
			Message: renameFileResponse.Message,
		}
	}

	return nil
}
