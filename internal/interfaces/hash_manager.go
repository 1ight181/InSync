package interfaces

import "insync/internal/domain"

type IHashManager interface {
	ResolveHash(fullPath string, fileMetadata domain.FileMetadata, getContent func(string) ([]byte, error)) (string, error)
	MarkDirty(fullPath string) error
}
