package server

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (gs *GrpcServer) CreateDir(ctx context.Context, request *insyncpb.CreateDirRequest) (*emptypb.Empty, error) {
	rootName := request.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	relativePath := request.GetRelativePath()
	validRelativePath, err := domain.NewPath(relativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	scopedPath, err := domain.NewScopedPath(validRootName, validRelativePath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	err = gs.fileUseCase.CreateDir(ctx, scopedPath)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
