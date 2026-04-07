package interfaces

import (
	"insync/internal/domain"
	"io"
)

type IFileManager interface {
	GetFileList(rootName string) ([]domain.FileEntry, error)
	DeleteFile(rootName string, relativePath string) error
	PutFile(rootName string, relativePath string, fileData io.Reader) error
	RenameFile(rootName string, oldRelativePath string, newRelativePath string) error
	GetFile(rootName string, relativePath string) (io.ReadCloser, error)
}
