package server

import (
	"context"

	"insync/internal/domain"
	insyncpb "insync/internal/transport/grpc/insyncpb"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (gs *GrpcServer) UpdateBaseSnapshot(ctx context.Context, request *insyncpb.UpdateBaseSnapshotRequest) (*emptypb.Empty, error) {
	rootName := request.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	err = gs.baseUseCase.UpdateBaseSnapshot(ctx, validRootName)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
