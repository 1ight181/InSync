package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"insync/internal/domain"
	ifaces "insync/internal/interfaces"
	"insync/internal/transport/grpc/insyncpb"
	"io"
	"log/slog"
	"net"
	"os"
	"sync/atomic"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GrpcServer struct {
	insyncpb.UnimplementedFileSyncServiceServer

	fileUseCase ifaces.IFileUseCase

	certPath   string
	keyPath    string
	caCertPath string

	networkType string
	address     string
	serviceName string

	chunkSizeInBytes int

	ctx context.Context

	logger    *slog.Logger
	loggerCtx context.Context

	server *grpc.Server

	healthServer *health.Server

	isStarted atomic.Bool
}

type GrpcServerOptions struct {
	FileUseCase ifaces.IFileUseCase

	CertPath   string
	KeyPath    string
	CaCertPath string

	NetworkType string
	Address     string
	ServiceName string

	ChunkSizeInBytes int

	Ctx context.Context

	Logger *slog.Logger
}

func NewGrpcServer(opts GrpcServerOptions) ifaces.IServer {
	if opts.FileUseCase == nil ||
		opts.CertPath == "" ||
		opts.KeyPath == "" ||
		opts.CaCertPath == "" ||

		opts.NetworkType == "" ||
		opts.Address == "" ||
		opts.ServiceName == "" ||

		opts.ChunkSizeInBytes <= 0 ||

		opts.Ctx == nil ||

		opts.Logger == nil {
		panic("Все поля GrpcServerOption должны быть заполнены")
	}
	loggerCtx := context.Background()
	return &GrpcServer{
		fileUseCase: opts.FileUseCase,

		certPath:   opts.CertPath,
		keyPath:    opts.KeyPath,
		caCertPath: opts.CaCertPath,

		networkType: opts.NetworkType,
		address:     opts.Address,
		serviceName: opts.ServiceName,

		chunkSizeInBytes: opts.ChunkSizeInBytes,

		ctx: opts.Ctx,

		logger:    opts.Logger,
		loggerCtx: loggerCtx,
	}
}

func (gs *GrpcServer) createTransportCreds() (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(gs.certPath, gs.keyPath)
	if err != nil {
		return nil, err
	}

	gs.logger.LogAttrs(
		gs.loggerCtx,
		slog.LevelDebug,
		"Сертификат сервера успешно загружен",
		slog.String("certPath", gs.certPath),
		slog.String("keyPath", gs.keyPath),
	)

	caCert, err := os.ReadFile(gs.caCertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	isOk := caCertPool.AppendCertsFromPEM(caCert)
	if !isOk {
		return nil, ErrFailedToAppendCa
	}

	gs.logger.LogAttrs(
		gs.loggerCtx,
		slog.LevelDebug,
		"CA сертификат успешно загружен для сервера",
		slog.String("caCertPath", gs.caCertPath),
	)

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}), nil
}

// Start блокирует выполнение, поэтому его нужно запускать в отдельной горутине. Он будет работать до тех пор, пока сервер не будет остановлен через метод Stop.
func (gs *GrpcServer) Start() (err error) {
	if !gs.isStarted.CompareAndSwap(false, true) {
		gs.logger.Warn("Попытка запустить сервер, который уже запущен")
		return ErrServerAlreadyStarted
	}
	defer func() {
		if err != nil {
			gs.isStarted.Store(false)
		}
	}()

	gs.logger.Info("Запуск gRPC сервера...")

	transportCreds, err := gs.createTransportCreds()
	if err != nil {
		return err
	}
	withTransportCreds := grpc.Creds(transportCreds)

	server := grpc.NewServer(withTransportCreds)
	gs.server = server
	insyncpb.RegisterFileSyncServiceServer(server, gs)

	healthServer := health.NewServer()
	gs.healthServer = healthServer
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	listener, err := net.Listen(gs.networkType, gs.address)
	if err != nil {
		return err
	}

	ctx := gs.ctx
	if err := ctx.Err(); err != nil {
		return err
	}

	go func() {
		defer gs.isStarted.Store(false)
		if err := server.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			gs.logger.LogAttrs(
				gs.loggerCtx,
				slog.LevelError,
				"Произошла ошибка при запуске gRPC сервера",
				slog.String("err", err.Error()),
			)
		}
	}()

	healthServer.SetServingStatus(gs.serviceName, grpc_health_v1.HealthCheckResponse_SERVING)

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
		return ErrServerAlreadyStopped
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

func (gs *GrpcServer) GetFileList(ctx context.Context, request *insyncpb.GetFileListRequest) (*insyncpb.GetFileListResponse, error) {
	rootName := request.GetRootName()

	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return nil, err
	}

	files, err := gs.fileUseCase.GetFileList(ctx, validRootName)
	if err != nil {
		return nil, err
	}

	pbFiles := make([]*insyncpb.FileMetadata, 0, len(files))
	for _, file := range files {
		pbFiles = append(pbFiles, DomainFileMetadataToPb(file))
	}

	response := &insyncpb.GetFileListResponse{
		Files: pbFiles,
	}

	return response, nil
}

func (gs *GrpcServer) GetFile(request *insyncpb.GetFileRequest, stream grpc.ServerStreamingServer[insyncpb.GetFileResponse]) error {
	ctx := stream.Context()

	rootName := request.GetRootName()
	relativePath := request.GetRelativePath()

	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return err
	}

	validRelativePath, err := domain.NewRelativePath(relativePath)
	if err != nil {
		return err
	}

	fileReader, err := gs.fileUseCase.GetFile(ctx, validRootName, validRelativePath)
	if err != nil {
		return err
	}
	defer fileReader.Close()

readLabel:
	for i := 0; ; i++ {
		buffer := make([]byte, gs.chunkSizeInBytes)

		readChan := make(chan struct {
			numberOfBytes int
			err           error
		})

		go func() {
			numberOfBytes, err := fileReader.Read(buffer)
			readChan <- struct {
				numberOfBytes int
				err           error
			}{
				numberOfBytes: numberOfBytes,
				err:           err,
			}
		}()

		select {
		case readResult := <-readChan:
			numberOfBytes := readResult.numberOfBytes
			err = readResult.err

			if err == io.EOF {
				break readLabel
			}
			if err != nil {
				return err
			}
			err = stream.Send(&insyncpb.GetFileResponse{
				Chunk: &insyncpb.FileChunk{
					Index: uint32(i),
					// обрезка буфера, чтобы не передавать нули на последней итерации
					Data: buffer[:numberOfBytes],
				},
			})
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (gs *GrpcServer) PutFile(stream grpc.ClientStreamingServer[insyncpb.PutFileRequest, insyncpb.PutFileResponse]) error {
	ctx := stream.Context()

	initRequest, err := stream.Recv()
	if err != nil {
		return err
	}

	initMessage := initRequest.GetInit()

	rootName := initMessage.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return err
	}

	relativePath := initMessage.GetRelativePath()
	validRelativePath, err := domain.NewRelativePath(relativePath)
	if err != nil {
		return err
	}

	pipeReader, pipeWriter := io.Pipe()

	err = gs.fileUseCase.PutFile(ctx, validRootName, validRelativePath, pipeReader)
	if err != nil {
		return err
	}
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		errChan := make(chan error)

		chunk := request.GetChunk()
		chunkData := chunk.GetData()

		go func() {
			_, err = pipeWriter.Write(chunkData)
			errChan <- err
		}()

		select {
		case writeErr := <-errChan:
			if writeErr != nil {
				if err := pipeWriter.CloseWithError(writeErr); errors.Is(err, io.ErrClosedPipe) {
					gs.logger.LogAttrs(
						gs.loggerCtx,
						slog.LevelWarn,
						"Попытка повторно закрыть пайп методом CloseWithError при выполнении PutFile, который уже был закрыт",
						slog.String("writeErr", writeErr.Error()),
					)
				}
				return writeErr
			}
		case <-ctx.Done():
			ctxErr := ctx.Err()
			if err := pipeWriter.CloseWithError(ctxErr); errors.Is(err, io.ErrClosedPipe) {
				gs.logger.LogAttrs(
					gs.loggerCtx,
					slog.LevelWarn,
					"Попытка повторно закрыть пайп методом CloseWithError при выполнении PutFile, который уже был закрыт",
					slog.String("ctxErr", ctxErr.Error()),
				)
			}
			return ctxErr
		}
	}

	return nil
}

func (gs *GrpcServer) DeleteFile(ctx context.Context, request *insyncpb.DeleteFileRequest) (*insyncpb.DeleteFileResponse, error) {
	rootName := request.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return &insyncpb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	relativePath := request.GetRelativePath()
	validRelativePath, err := domain.NewRelativePath(relativePath)
	if err != nil {
		return &insyncpb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	err = gs.fileUseCase.DeleteFile(ctx, validRootName, validRelativePath)
	if err != nil {
		return &insyncpb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &insyncpb.DeleteFileResponse{
		Success: true,
	}, nil
}

func (gs *GrpcServer) RenameFile(ctx context.Context, request *insyncpb.RenameFileRequest) (*insyncpb.RenameFileResponse, error) {
	rootName := request.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return &insyncpb.RenameFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	oldRelativePath := request.GetOldRelativePath()
	validOldRelativePath, err := domain.NewRelativePath(oldRelativePath)
	if err != nil {
		return &insyncpb.RenameFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	newRelativePath := request.GetNewRelativePath()
	validNewRelativePath, err := domain.NewRelativePath(newRelativePath)
	if err != nil {
		return &insyncpb.RenameFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	err = gs.fileUseCase.RenameFile(ctx, validRootName, validOldRelativePath, validNewRelativePath)
	if err != nil {
		return &insyncpb.RenameFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &insyncpb.RenameFileResponse{
		Success: true,
	}, nil
}
