package server

import (
	"context"

	"insync/internal/domain"
	insyncpb "insync/internal/transport/grpc/insyncpb"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (gs *GrpcServer) SetBaseSnapshot(ctx context.Context, request *insyncpb.SetBaseSnapshotRequest) (*emptypb.Empty, error) {
	rootName := request.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	rawSnapshot := request.GetSnapshot()
	snapshot := pbSnapshotToDomain(rawSnapshot)

	err = gs.baseUseCase.SetBaseSnapshot(ctx, snapshot, validRootName)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
