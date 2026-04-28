package client

import (
	"context"
	"insync/internal/domain"

	insyncpb "insync/internal/transport/grpc/insyncpb"
)

func (gc *GrpcClient) CreateDir(ctx context.Context, scopedPath domain.ScopedPath) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка удалить файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	createDirRequest := &insyncpb.CreateDirRequest{
		RootName:     scopedPath.Root.String(),
		RelativePath: scopedPath.Path.String(),
	}

	_, err := gc.client.CreateDir(ctx, createDirRequest)
	if err != nil {
		return err
	}

	return nil
}
