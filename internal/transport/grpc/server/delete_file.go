package server

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func (gs *GrpcServer) DeleteFile(ctx context.Context, request *insyncpb.DeleteFileRequest) (*insyncpb.DeleteFileResponse, error) {
	rootName := request.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return &insyncpb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	relativePath := request.GetRelativePath()
	validRelativePath, err := domain.NewRelativePath(relativePath)
	if err != nil {
		return &insyncpb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	err = gs.fileUseCase.DeleteFile(ctx, validRootName, validRelativePath)
	if err != nil {
		return &insyncpb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &insyncpb.DeleteFileResponse{
		Success: true,
	}, nil
}