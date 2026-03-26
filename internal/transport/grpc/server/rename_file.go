package server

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func (gs *GrpcServer) RenameFile(ctx context.Context, request *insyncpb.RenameFileRequest) (*insyncpb.RenameFileResponse, error) {
	rootName := request.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return &insyncpb.RenameFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	oldRelativePath := request.GetOldRelativePath()
	validOldRelativePath, err := domain.NewRelativePath(oldRelativePath)
	if err != nil {
		return &insyncpb.RenameFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	newRelativePath := request.GetNewRelativePath()
	validNewRelativePath, err := domain.NewRelativePath(newRelativePath)
	if err != nil {
		return &insyncpb.RenameFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	err = gs.fileUseCase.RenameFile(ctx, validRootName, validOldRelativePath, validNewRelativePath)
	if err != nil {
		return &insyncpb.RenameFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &insyncpb.RenameFileResponse{
		Success: true,
	}, nil
}