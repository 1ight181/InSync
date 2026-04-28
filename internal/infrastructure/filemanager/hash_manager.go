package filemanager

import (
	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"
)

type IHashManager interface {
	ResolveHash(resourceContent cont.ResourceContent, rootName domain.RootName) (string, error)
	MarkDirty(scopedPath domain.ScopedPath) error
	ResolveWithForceRecalc(resourceContent cont.ResourceContent) (string, error)
}
