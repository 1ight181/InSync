package client

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"

	"go.uber.org/multierr"
)

func (gc *GrpcClient) GetFileList(ctx context.Context, rootName domain.RootName) ([]domain.FileEntry, error) {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда клиент не запущен")
		return nil, ErrClientNotStarted
	}

	getFileListResponse, err := gc.client.GetFileList(ctx, &insyncpb.GetFileListRequest{
		RootName: rootName.String(),
	})
	if err != nil {
		return nil, err
	}

	var resultErr error
	fileList := make([]domain.FileEntry, 0, len(getFileListResponse.Files))
	for _, file := range getFileListResponse.Files {
		fileMetadata := domain.FileMetadata{
			ModifiedUnix: file.ModifiedUnix,
			SizeBytes:    file.SizeBytes,
			IsDirectory:  file.IsDirectory,
		}

		fileInfo, err := domain.NewFileInfo(fileMetadata, file.Hash)
		if err != nil {
			resultErr = multierr.Append(resultErr, err)
			continue
		}

		filePath, err := domain.NewPath(file.RelativePath)
		if err != nil {
			resultErr = multierr.Append(resultErr, err)
			continue
		}

		fileEntry, err := domain.NewFileEntry(filePath, fileInfo)
		if err != nil {
			resultErr = multierr.Append(resultErr, err)
			continue
		}

		fileList = append(fileList, fileEntry)
	}

	if resultErr != nil {
		return nil, resultErr
	}

	return fileList, nil
}
