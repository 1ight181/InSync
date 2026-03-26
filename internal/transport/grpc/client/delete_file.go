package client

import (
	"context"
	"insync/internal/transport/grpc/insyncpb"
)

func (gc *GrpcClient) DeleteFile(ctx context.Context, rootName, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка удалить файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	deleteFileRequest := &insyncpb.DeleteFileRequest{
		RootName:     rootName,
		RelativePath: relativePath,
	}

	_, err := gc.client.DeleteFile(ctx, deleteFileRequest)
	if err != nil {
		return err
	}

	return nil
}
