package root

import "insync/internal/domain"

type IRootResolverRepository interface {
	GetRoots() (map[domain.RootName]domain.Path, error)
	AddRoot(rootName domain.RootName, rootPath domain.Path) error
	RemoveRoot(rootName domain.RootName) error
}
