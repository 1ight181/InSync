package client

import (
	"context"
	"insync/internal/transport/grpc/insyncpb"
)

func (gc *GrpcClient) RenameFile(ctx context.Context, rootName, fileUuid, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка переименовать файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	renameFileRequest := &insyncpb.RenameFileRequest{
		RootName:        rootName,
		NewRelativePath: relativePath,
	}

	renameFileResponse, err := gc.client.RenameFile(ctx, renameFileRequest)
	if err != nil {
		return err
	}
	if !renameFileResponse.Success {
		return RenameFileFailedError{
			Message: renameFileResponse.Message,
		}
	}

	return nil
}
