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
	validOldRelativePath, err := domain.NewPath(oldRelativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	newRelativePath := request.GetNewRelativePath()
	validNewRelativePath, err := domain.NewPath(newRelativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	oldScopedPath, err := domain.NewScopedPath(validRootName, validOldRelativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	newScopedPath, err := domain.NewScopedPath(validRootName, validNewRelativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	err = gs.fileUseCase.RenameFile(ctx, oldScopedPath, newScopedPath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
