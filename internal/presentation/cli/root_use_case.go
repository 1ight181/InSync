package cli

import "insync/internal/domain"

type IRootUseCase interface {
	AddRoot(rootName domain.RootName, relativePath domain.Path) error
	RemoveRoot(rootName domain.RootName) error
	GetRoots() (map[domain.RootName]domain.Path, error)
}
