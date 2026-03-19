package interfaces

import (
	"context"
	models "insync/internal/domain"
	"io"
)

type IClient interface {
	Start() error
	Stop() error
	GetFileList(ctx context.Context, rootName string) ([]models.FileMetadata, error)
	GetFile(ctx context.Context, rootName, relativePath string) (*io.PipeReader, error)
	PutFile(ctx context.Context, file io.Reader, rootName, relativePath string) error
	DeleteFile(ctx context.Context, rootName, relativePath string) error
	RenameFile(ctx context.Context, rootName, fileUuid, relativePath string) error
}
