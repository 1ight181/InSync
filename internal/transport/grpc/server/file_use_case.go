package server

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IFileUseCase interface {
	GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.SnapshotWithMetadata, error)
	DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) error
	PutFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path, file io.Reader) error
	RenameFile(ctx context.Context, rootName domain.RootName, oldRelativePath, newRelativePath domain.Path) error
	GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) (io.ReadCloser, error)
}
