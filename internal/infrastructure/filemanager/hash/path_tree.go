package hash

import "insync/internal/domain"

type IPathTreeReader interface {
	GetChildren(fullPath domain.Path) ([]domain.Path, error)
	GetParents(fullPath domain.Path) ([]domain.Path, error)
}
