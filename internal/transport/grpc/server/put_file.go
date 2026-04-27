package server

import (
	"bytes"
	"errors"
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
	"io"

	"google.golang.org/protobuf/types/known/emptypb"

	grpc "google.golang.org/grpc"
)

func (gs *GrpcServer) PutFile(stream grpc.ClientStreamingServer[insyncpb.PutFileRequest, emptypb.Empty]) error {
	ctx := stream.Context()

	initRequest, err := stream.Recv()
	if err != nil {
		return err
	}

	initMessage := initRequest.GetInit()
	if initMessage == nil {
		return errors.New("missing init message")
	}

	rootName := initMessage.GetRootName()
	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return err
	}

	relativePath := initMessage.GetRelativePath()
	validRelativePath, err := domain.NewPath(relativePath)
	if err != nil {
		return err
	}

	scopedPath, err := domain.NewScopedPath(validRootName, validRelativePath)
	if err != nil {
		return err
	}

	var content bytes.Buffer

	for {
		request, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		chunk := request.GetChunk()
		if chunk == nil {
			return errors.New("missing chunk message")
		}

		if _, err := content.Write(chunk.GetData()); err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return err
		}
	}

	reader := bytes.NewReader(content.Bytes())

	if err := gs.fileUseCase.PutFile(ctx, scopedPath, reader); err != nil {
		return err
	}

	return stream.SendAndClose(&emptypb.Empty{})
}
