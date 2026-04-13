package filemanager

import (
	cont "insync/internal/infrastructure/filemanager/content"
)

type IHashManager interface {
	ResolveHash(resourceContent cont.ResourceContent, fullPath string) (string, error)
	MarkDirty(fullPath string) error
}
