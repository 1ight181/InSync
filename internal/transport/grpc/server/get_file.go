package server

import (
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
	"io"

	grpc "google.golang.org/grpc"
)

func (gs *GrpcServer) GetFile(request *insyncpb.GetFileRequest, stream grpc.ServerStreamingServer[insyncpb.GetFileResponse]) error {
	ctx := stream.Context()

	rootName := request.GetRootName()
	relativePath := request.GetRelativePath()

	validRootName, err := domain.NewRootName(rootName)
	if err != nil {
		return err
	}

	validRelativePath, err := domain.NewPath(relativePath)
	if err != nil {
		return err
	}

	scopedPath, err := domain.NewScopedPath(validRootName, validRelativePath)
	if err != nil {
		return err
	}

	fileReader, err := gs.fileUseCase.GetFile(ctx, scopedPath)
	if err != nil {
		return err
	}
	defer fileReader.Close()

readLabel:
	for i := 0; ; i++ {
		buffer := make([]byte, gs.chunkSizeInBytes)

		readChan := make(chan struct {
			numberOfBytes int
			err           error
		})

		go func() {
			numberOfBytes, err := fileReader.Read(buffer)
			readChan <- struct {
				numberOfBytes int
				err           error
			}{
				numberOfBytes: numberOfBytes,
				err:           err,
			}
		}()

		select {
		case readResult := <-readChan:
			numberOfBytes := readResult.numberOfBytes
			err = readResult.err

			if err == io.EOF {
				break readLabel
			}
			if err != nil {
				return err
			}
			err = stream.Send(&insyncpb.GetFileResponse{
				Chunk: &insyncpb.FileChunk{
					Index: uint64(i),
					// обрезка буфера, чтобы не передавать нули на последней итерации
					Data: buffer[:numberOfBytes],
				},
			})
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
