package interfaces

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IFileManager interface {
	GetFileList(ctx context.Context, rootName string) ([]domain.FileEntry, error)
	DeleteFile(ctx context.Context, rootName string, relativePath string) error
	PutFile(ctx context.Context, rootName string, relativePath string, fileData io.Reader) error
	RenameFile(ctx context.Context, rootName string, oldRelativePath string, newRelativePath string) error
	GetFile(ctx context.Context, rootName string, relativePath string) (io.ReadCloser, error)
}
