package interfaces

import (
	"insync/internal/domain"
	"io"
)

type IFileUseCase interface {
	GetFileList(rootName string) ([]domain.FileMetadata, error)
	DeleteFile(rootName, relativePath string) error
	PutFile(rootName, relativePath string, file io.Reader) error
	RenameFile(rootName, oldRelativePath, newRelativePath string) error
	GetFile(rootName, relativePath string) (*io.PipeReader, error)
}
