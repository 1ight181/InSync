package root

import "insync/internal/domain"

type IRootRegistrar interface {
	AddRoot(rootName domain.RootName, rootPath domain.Path) error
	RemoveRoot(rootName domain.RootName) error
	GetRoots() map[domain.RootName]domain.Path
}
