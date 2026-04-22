package server

import (
	"errors"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
	"io"
	"log/slog"

	"google.golang.org/protobuf/types/known/emptypb"

	grpc "google.golang.org/grpc"
)

func (gs *GrpcServer) PutFile(stream grpc.ClientStreamingServer[insyncpb.PutFileRequest, emptypb.Empty]) error {
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
	validRelativePath, err := domain.NewPath(relativePath)
	if err != nil {
		return err
	}

	scopedPath, err := domain.NewScopedPath(validRootName, validRelativePath)
	if err != nil {
		return err
	}

	pipeReader, pipeWriter := io.Pipe()

	err = gs.fileUseCase.PutFile(ctx, scopedPath, pipeReader)
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

	if err := pipeWriter.Close(); err != nil {
		return err
	}

	return stream.SendAndClose(
		&emptypb.Empty{},
	)
}
