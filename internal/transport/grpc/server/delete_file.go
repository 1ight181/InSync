package server

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (gs *GrpcServer) DeleteFile(ctx context.Context, request *insyncpb.DeleteFileRequest) (*emptypb.Empty, error) {
	rootName := request.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	relativePath := request.GetRelativePath()
	validRelativePath, err := domain.NewRelativePath(relativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	err = gs.fileUseCase.DeleteFile(ctx, validRootName, validRelativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
