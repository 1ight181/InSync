package client

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func (gc *GrpcClient) RenameFile(ctx context.Context, fileUuid string, rootName domain.RootName, oldRelativePath domain.Path, newRelativePath domain.Path) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка переименовать файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	renameFileRequest := &insyncpb.RenameFileRequest{
		RootName:        rootName.String(),
		OldRelativePath: oldRelativePath.String(),
		NewRelativePath: newRelativePath.String(),
	}

	_, err := gc.client.RenameFile(ctx, renameFileRequest)
	if err != nil {
		return err
	}

	return nil
}
