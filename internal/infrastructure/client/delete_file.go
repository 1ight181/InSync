package client

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func (gc *GrpcClient) DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка удалить файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	deleteFileRequest := &insyncpb.DeleteFileRequest{
		RootName:     rootName.String(),
		RelativePath: relativePath.String(),
	}

	_, err := gc.client.DeleteFile(ctx, deleteFileRequest)
	if err != nil {
		return err
	}

	return nil
}
