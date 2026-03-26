package server

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (gs *GrpcServer) RenameFile(ctx context.Context, request *insyncpb.RenameFileRequest) (*emptypb.Empty, error) {
	rootName := request.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	oldRelativePath := request.GetOldRelativePath()
	validOldRelativePath, err := domain.NewRelativePath(oldRelativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	newRelativePath := request.GetNewRelativePath()
	validNewRelativePath, err := domain.NewRelativePath(newRelativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	err = gs.fileUseCase.RenameFile(ctx, validRootName, validOldRelativePath, validNewRelativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
