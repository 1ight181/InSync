package server

import (
	"context"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func (gs *GrpcServer) GetFileList(ctx context.Context, request *insyncpb.GetFileListRequest) (*insyncpb.GetFileListResponse, error) {
	rootName := request.GetRootName()

	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return nil, err
	}

	files, err := gs.fileUseCase.GetFileList(ctx, validRootName)
	if err != nil {
		return nil, err
	}

	pbFiles := make([]*insyncpb.FileEntry, 0, len(files))
	for _, file := range files {
		pbFiles = append(pbFiles, DomainFileEntryToPb(file))
	}

	response := &insyncpb.GetFileListResponse{
		Files: pbFiles,
	}

	return response, nil
}
