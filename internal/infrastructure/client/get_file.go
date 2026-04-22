package client

import (
	"context"
	"errors"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
	"io"
	"log/slog"
	"runtime/debug"
)

func (gc *GrpcClient) GetFile(ctx context.Context, scopedPath domain.ScopedPath) (io.ReadCloser, error) {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка получить файл, когда клиент не запущен")
		return nil, ErrClientNotStarted
	}

	getFileRequest := &insyncpb.GetFileRequest{
		RootName:     scopedPath.Root.String(),
		RelativePath: scopedPath.Path.String(),
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
					getFileResponseStream.CloseSend()
					gc.logger.LogAttrs(
						gc.loggerCtx,
						slog.LevelError,
						"Ошибка при записи данных в пайп методом Write при выполнении GetFile",
						slog.String("error", writeErr.Error()),
					)
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
				getFileResponseStream.CloseSend()

				if closeErr := pipeWriter.CloseWithError(ctx.Err()); errors.Is(closeErr, io.ErrClosedPipe) {
					gc.logger.Warn(
						"Попытка повторно закрыть пайп методом CloseWithError при выполнении GetFile, который уже был закрыт",
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
