package client

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func (gc *GrpcClient) GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда клиент не запущен")
		return domain.Snapshot{}, ErrClientNotStarted
	}

	getSnapshotResponse, err := gc.client.GetSnapshot(ctx, &insyncpb.GetSnapshotRequest{
		RootName: rootName.String(),
	})
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot, err := pbSnapshotToDomain(getSnapshotResponse.GetSnapshot())
	if err != nil {
		return domain.Snapshot{}, err
	}

	return snapshot, nil
}
