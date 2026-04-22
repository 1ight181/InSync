package client

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func (gc *GrpcClient) DeleteFile(ctx context.Context, scopedPath domain.ScopedPath) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка удалить файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	deleteFileRequest := &insyncpb.DeleteFileRequest{
		RootName:     scopedPath.Root.String(),
		RelativePath: scopedPath.Path.String(),
	}

	_, err := gc.client.DeleteFile(ctx, deleteFileRequest)
	if err != nil {
		return err
	}

	return nil
}
