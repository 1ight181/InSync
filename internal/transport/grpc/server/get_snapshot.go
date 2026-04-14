package server

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func (gs *GrpcServer) GetSnapshot(ctx context.Context, request *insyncpb.GetSnapshotRequest) (*insyncpb.GetSnapshotResponse, error) {
	rootName := request.GetRootName()

	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return nil, err
	}

	snapshot, err := gs.fileUseCase.GetSnapshot(ctx, validRootName)
	if err != nil {
		return nil, err
	}

	pbSnapshot := domainSnapshotToPb(snapshot)

	response := &insyncpb.GetSnapshotResponse{
		Snapshot: pbSnapshot,
	}

	return response, nil
}
