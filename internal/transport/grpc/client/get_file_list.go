package client

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

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
