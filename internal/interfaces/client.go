package interfaces

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IClient interface {
	GetFileList(ctx context.Context, rootName domain.RootName) ([]domain.FileEntry, error)
	// Причина использования именно ReadCloser вместо io.Reader, так как ReadCloser позволяет закрыть соединение с сервером
	// Это позволяет закрыть соединение, даже если реализация io.Reader не закрывает соединение по контексту, игнорируя его
	GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) (io.ReadCloser, error)
	PutFile(ctx context.Context, file io.Reader, rootName domain.RootName, relativePath domain.Path) error
	DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) error
	RenameFile(ctx context.Context, fileUuid string, rootName domain.RootName, oldRelativePath domain.Path, newRelativePath domain.Path) error
}
