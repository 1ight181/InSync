package interfaces

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IFileUseCase interface {
	GetFileList(ctx context.Context, rootName string) ([]domain.FileMetadata, error)
	DeleteFile(ctx context.Context, rootName, relativePath string) error
	PutFile(ctx context.Context, rootName, relativePath string, file io.Reader) error
	RenameFile(ctx context.Context, rootName, oldRelativePath, newRelativePath string) error
	GetFile(ctx context.Context, rootName, relativePath string) (*io.ReadCloser, error)
}
