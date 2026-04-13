package filemanager

import (
	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"
)

type IHashManager interface {
	ResolveHash(resourceContent cont.ResourceContent, fileMetadata domain.FileMetadata) (string, error)
	MarkDirty(fullPath string) error
}
