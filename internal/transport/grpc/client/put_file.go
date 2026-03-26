package client

import (
	"context"
	"insync/internal/transport/grpc/insyncpb"
	"io"
)

func (gc *GrpcClient) PutFile(ctx context.Context, file io.Reader, rootName, relativePath string) error {
	if !gc.isStarted.Load() {
		gc.logger.Warn("Попытка отправить файл, когда клиент не запущен")
		return ErrClientNotStarted
	}

	putFileStream, err := gc.client.PutFile(ctx)
	if err != nil {
		return err
	}

	initMessage := &insyncpb.PutFileRequest{
		Payload: &insyncpb.PutFileRequest_Init{
			Init: &insyncpb.PutFileInit{
				RootName:     rootName,
				RelativePath: relativePath,
			},
		},
	}

	err = putFileStream.Send(initMessage)
	if err != nil {
		return err
	}

	buffer := make([]byte, gc.chunkSizeInBytes)

readLabel:
	for i := 0; ; i++ {
		readChan := make(chan struct {
			numberOfBytes int
			err           error
		})

		go func() {
			numberOfBytes, err := file.Read(buffer)
			readChan <- struct {
				numberOfBytes int
				err           error
			}{
				numberOfBytes: numberOfBytes,
				err:           err,
			}
		}()

		select {
		case <-ctx.Done():
			putFileStream.CloseSend()
			return ctx.Err()
		case readResult := <-readChan:
			numberOfBytes := readResult.numberOfBytes
			err = readResult.err
			if numberOfBytes == 0 && err == io.EOF {
				break readLabel
			}
			if err != nil {
				putFileStream.CloseSend()
				return err
			}

			chunkMessage := &insyncpb.PutFileRequest{
				Payload: &insyncpb.PutFileRequest_Chunk{
					Chunk: &insyncpb.FileChunk{
						Index: uint32(i),
						Data:  buffer[:numberOfBytes],
					},
				},
			}

			err = putFileStream.Send(chunkMessage)
			if err != nil {
				return err
			}
		}

	}

	putFileResponse, err := putFileStream.CloseAndRecv()
	if err != nil {
		return err
	}
	if !putFileResponse.Success {
		return PutFileFailedError{
			Message: putFileResponse.Message,
		}
	}

	return nil
}
