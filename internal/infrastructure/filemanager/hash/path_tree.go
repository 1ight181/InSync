package hash

import "insync/internal/domain"

type IPathTreeReader interface {
	GetParents(scopedPath domain.ScopedPath) ([]domain.ScopedPath, error)
}
