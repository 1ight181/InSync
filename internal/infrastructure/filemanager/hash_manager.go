package filemanager

import "insync/internal/domain"

type IHashManager interface {
	ResolveHash(resourceContent domain.ResourceContent, fileMetadata domain.FileMetadata) (string, error)
	MarkDirty(fullPath string) error
}
