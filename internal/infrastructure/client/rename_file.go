package client

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func (gc *GrpcClient) RenameFile(ctx context.Context, oldScopedPath domain.ScopedPath, newScopedPath domain.ScopedPath) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка переименовать файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	renameFileRequest := &insyncpb.RenameFileRequest{
		RootName:        oldScopedPath.Root.String(),
		OldRelativePath: oldScopedPath.Path.String(),
		NewRelativePath: newScopedPath.Path.String(),
	}

	_, err := gc.client.RenameFile(ctx, renameFileRequest)
	if err != nil {
		return err
	}

	return nil
}
