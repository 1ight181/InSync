package interfaces

import (
	"context"
	"insync/internal/domain"
	"io"
)

type IClient interface {
	Start() error
	Stop() error
	GetFileList(ctx context.Context, rootName string) ([]domain.FileInfo, error)
	// Причина использования именно ReadCloser вместо io.Reader, так как ReadCloser позволяет закрыть соединение с сервером
	// Это позволяет закрыть соединение, даже если реализация io.Reader не закрывает соединение по контексту, игнорируя его
	GetFile(ctx context.Context, rootName, relativePath string) (io.ReadCloser, error)
	PutFile(ctx context.Context, file io.Reader, rootName, relativePath string) error
	DeleteFile(ctx context.Context, rootName, relativePath string) error
	RenameFile(ctx context.Context, rootName, fileUuid, relativePath string) error
}
