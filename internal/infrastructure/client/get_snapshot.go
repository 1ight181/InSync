package client

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func (gc *GrpcClient) GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.SnapshotWithMetadata, error) {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка получить список файлов, когда клиент не запущен")
		return domain.SnapshotWithMetadata{}, ErrClientNotStarted
	}

	getSnapshotResponse, err := gc.client.GetSnapshot(ctx, &insyncpb.GetSnapshotRequest{
		RootName: rootName.String(),
	})
	if err != nil {
		return domain.SnapshotWithMetadata{}, err
	}

	snapshot, err := pbSnapshotToDomain(getSnapshotResponse.GetSnapshot())
	if err != nil {
		return domain.SnapshotWithMetadata{}, err
	}

	return snapshot, nil
}
