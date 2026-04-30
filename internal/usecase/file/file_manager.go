package file

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IFileManager interface {
	GetSnapshot(ctx context.Context, rootName domain.RootName, baseSnapshot *domain.BaseSnapshot) (domain.Snapshot, error)
	DeleteFile(ctx context.Context, scopedPath domain.ScopedPath) error
	PutFile(ctx context.Context, scopedPath domain.ScopedPath, fileData io.Reader) error
	RenameFile(ctx context.Context, oldScopedPath domain.ScopedPath, newScopedPath domain.ScopedPath) error
	GetFile(ctx context.Context, scopedPath domain.ScopedPath) (io.ReadCloser, error)
	CreateDir(ctx context.Context, scopedPath domain.ScopedPath) error
}
