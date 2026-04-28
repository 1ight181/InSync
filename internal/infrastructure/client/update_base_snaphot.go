package client

import (
	"context"
	"insync/internal/domain"
	insyncpb "insync/internal/transport/grpc/insyncpb"
)

func (gc *GrpcClient) UpdateBaseSnapshot(ctx context.Context, rootName domain.RootName) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка получить файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	updateSnapshotRequest := &insyncpb.UpdateBaseSnapshotRequest{
		RootName: rootName.String(),
	}

	_, err := gc.client.UpdateBaseSnapshot(ctx, updateSnapshotRequest)
	if err != nil {
		return err
	}

	return nil
}
