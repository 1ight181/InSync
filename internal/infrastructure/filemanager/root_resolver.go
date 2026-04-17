package filemanager

import "insync/internal/domain"

type IRootResolver interface {
	ResolveRoot(rootName domain.RootName, relativePath domain.Path) (domain.Path, error)
}
