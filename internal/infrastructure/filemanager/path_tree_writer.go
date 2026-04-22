package filemanager

import "insync/internal/domain"

type IPathTreeWriter interface {
	AddPath(scopedPath domain.ScopedPath) error
	RemovePath(scopedPath domain.ScopedPath) error
}
