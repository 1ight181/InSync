package local

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IFileManager interface {
	RenameFile(ctx context.Context, oldScopedPath domain.ScopedPath, newScopedPath domain.ScopedPath) error
	DeleteFile(ctx context.Context, scopedPath domain.ScopedPath) error
	PutFile(ctx context.Context, scopedPath domain.ScopedPath, content io.Reader) error
	GetFile(ctx context.Context, scopedPath domain.ScopedPath) (io.ReadCloser, error)
}
