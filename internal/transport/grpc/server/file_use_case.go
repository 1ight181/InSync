package server

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IFileUseCase interface {
	GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error)
	DeleteFile(ctx context.Context, scopedPath domain.ScopedPath) error
	PutFile(ctx context.Context, scopedPath domain.ScopedPath, file io.Reader) error
	RenameFile(ctx context.Context, oldScopedPath domain.ScopedPath, newScopedPath domain.ScopedPath) error
	GetFile(ctx context.Context, scopedPath domain.ScopedPath) (io.ReadCloser, error)
	CreateDir(ctx context.Context, scopedPath domain.ScopedPath) error
}
