package root

import (
	"insync/internal/domain"
	"path/filepath"
)

type RootResolver struct {
	rootMap map[domain.RootName]domain.Path
}

func NewRootResolver() *RootResolver {
	return &RootResolver{
		rootMap: make(map[domain.RootName]domain.Path),
	}
}

func (p *RootResolver) ResolveRoot(rootName domain.RootName, relativePath domain.Path) (string, error) {
	rootPath, ok := p.rootMap[rootName]
	if !ok {
		return "", RootNotFoundError{
			RootName: rootName,
		}
	}

	return filepath.Join(rootPath.String(), relativePath.String()), nil
}

func (p *RootResolver) AddRoot(rootName domain.RootName, rootPath domain.Path) {
	p.rootMap[rootName] = rootPath
}

func (p *RootResolver) RemoveRoot(rootName domain.RootName) {
	delete(p.rootMap, rootName)
}

func (p *RootResolver) GetRoots() []domain.RootName {
	roots := make([]domain.RootName, 0, len(p.rootMap))
	for rootName := range p.rootMap {
		roots = append(roots, rootName)
	}
	return roots
}
