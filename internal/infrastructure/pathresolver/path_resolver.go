package interfaces

import (
	interfaces "insync/internal/interfaces"
	"path/filepath"
)

type PathResolver struct {
	rootCache interfaces.IRootCache
}

func NewPathResolver(rootCache interfaces.IRootCache) interfaces.IPathResolver {
	return &PathResolver{rootCache: rootCache}
}

func (p *PathResolver) ResolvePath(rootName string, relativePath string) (string, error) {
	rootPath, err := p.rootCache.GetRootCache(rootName)
	if err != nil {
		return "", err
	}

	return filepath.Join(rootPath, relativePath), nil
}
