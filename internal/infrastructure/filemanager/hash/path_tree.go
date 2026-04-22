package hash

import "insync/internal/domain"

type IPathTreeReader interface {
	GetParents(fullPath domain.Path) ([]domain.Path, error)
}
