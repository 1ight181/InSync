package cli

import "insync/internal/domain"

type IRootUseCase interface {
	AddRoot(rootName domain.RootName, relativePath domain.Path)
	RemoveRoot(rootName domain.RootName)
	GetRoots() []domain.RootName
}
