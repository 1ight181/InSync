package interfaces

import "insync/internal/domain"

type IFileManager interface {
	GetFileList(rootName string) ([]domain.FileMetadata, error)
	DeleteFile(rootName string, relativePath string) error
	PutFile(rootName string, relativePath string, fileData []byte) error
	RenameFile(rootName string, oldRelativePath string, newRelativePath string) error
	GetFile(rootName string, relativePath string) ([]byte, error)
}
