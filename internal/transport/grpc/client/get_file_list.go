package client

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
	"log/slog"

	"go.uber.org/multierr"
)

func (gc *GrpcClient) GetFileList(ctx context.Context, rootName string) ([]domain.FileInfo, error) {
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

	var resultErr error
	fileList := make([]domain.FileInfo, 0, len(getFileListResponse.Files))
	for _, file := range getFileListResponse.Files {
		fileMetadata := domain.NewFileMetadata(
			file.ModifiedUnix,
			file.SizeBytes,
			file.IsDirectory,
			file.Hash,
		)

		fileInfo, err := domain.NewFileInfo(
			file.RootName,
			file.RelativePath,
			fileMetadata,
		)
		if err != nil {
			gc.logger.LogAttrs(
				gc.loggerCtx,
				slog.LevelWarn,
				"Получен некоректный fileMetadata от сервера, пропуск файла",
				slog.String("rootName", file.RootName),
				slog.String("relativePath", file.RelativePath),
				slog.String("hash", file.Hash),
			)
			resultErr = multierr.Append(resultErr, err)
			continue
		}

		fileList = append(fileList, fileInfo)
	}

	if resultErr != nil {
		return nil, resultErr
	}

	return fileList, nil
}
