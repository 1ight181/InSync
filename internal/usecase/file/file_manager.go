package file

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IFileManager interface {
	GetFileList(ctx context.Context, rootName domain.RootName) ([]domain.FileEntry, error)
	DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath) error
	PutFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath, fileData io.Reader) error
	RenameFile(ctx context.Context, rootName domain.RootName, oldRelativePath domain.RelativePath, newRelativePath domain.RelativePath) error
	GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath) (io.ReadCloser, error)
}
