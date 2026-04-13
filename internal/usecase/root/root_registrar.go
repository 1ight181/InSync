package root

import "insync/internal/domain"

type IRootRegistrar interface {
	AddRoot(rootName domain.RootName, rootPath domain.Path)
	RemoveRoot(rootName domain.RootName)
	GetRoots() []domain.RootName
}
