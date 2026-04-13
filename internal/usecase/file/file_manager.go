package file

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IFileManager interface {
	GetFileList(ctx context.Context, rootName domain.RootName) ([]domain.FileEntry, error)
	DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) error
	PutFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path, fileData io.Reader) error
	RenameFile(ctx context.Context, rootName domain.RootName, oldRelativePath domain.Path, newRelativePath domain.Path) error
	GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) (io.ReadCloser, error)
}
