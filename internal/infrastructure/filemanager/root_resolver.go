package filemanager

import "insync/internal/domain"

type IRootResolver interface {
	ResolveRoot(scopedPath domain.ScopedPath) (domain.Path, error)
}
