package interfaces

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IFileUseCase interface {
	GetFileList(ctx context.Context, rootName domain.RootName) ([]domain.FileMetadata, error)
	DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath) error
	PutFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath, file io.Reader) error
	RenameFile(ctx context.Context, rootName domain.RootName, oldRelativePath, newRelativePath domain.RelativePath) error
	GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath) (io.ReadCloser, error)
}
