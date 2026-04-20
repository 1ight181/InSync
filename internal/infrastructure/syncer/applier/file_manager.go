package local

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IFileManager interface {
	RenameFile(ctx context.Context, rootName domain.RootName, oldPath domain.Path, newPath domain.Path) error
	DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) error
	PutFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path, content io.Reader) error
	GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) (io.ReadCloser, error)
}
