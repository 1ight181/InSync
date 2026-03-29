package interfaces

import (
	interfaces "insync/internal/interfaces"
	"path/filepath"
)

type RootResolver struct {
	rootCache interfaces.IRootCache
}

func NewRootResolver(rootCache interfaces.IRootCache) interfaces.IRootResolver {
	return &RootResolver{rootCache: rootCache}
}

func (p *RootResolver) ResolveRoot(rootName string, relativePath string) (string, error) {
	rootPath, err := p.rootCache.GetRootCache(rootName)
	if err != nil {
		return "", err
	}

	return filepath.Join(rootPath, relativePath), nil
}
