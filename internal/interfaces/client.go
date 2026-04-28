package interfaces

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IClient interface {
	GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error)
	// Причина использования именно ReadCloser вместо io.Reader, так как ReadCloser позволяет закрыть соединение с сервером
	// Это позволяет закрыть соединение, даже если реализация io.Reader не закрывает соединение по контексту, игнорируя его
	GetFile(ctx context.Context, scopedPath domain.ScopedPath) (io.ReadCloser, error)
	PutFile(ctx context.Context, file io.Reader, scopedPath domain.ScopedPath) error
	DeleteFile(ctx context.Context, scopedPath domain.ScopedPath) error
	RenameFile(ctx context.Context, oldScopedPath domain.ScopedPath, newScopedPath domain.ScopedPath) error
	CreateDir(ctx context.Context, scopedPath domain.ScopedPath) error
	UpdateBaseSnapshot(ctx context.Context, rootName domain.RootName) error
}
