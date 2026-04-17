package filemanager

import (
	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"
)

type IHashManager interface {
	ResolveHash(resourceContent cont.ResourceContent, fullPath domain.Path) (string, error)
	MarkDirty(fullPath domain.Path) error
}
