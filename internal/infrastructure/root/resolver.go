package root

import (
	"insync/internal/domain"
)

type RootResolver struct {
	rootMap map[domain.RootName]domain.Path
}

func NewRootResolver() *RootResolver {
	return &RootResolver{
		rootMap: make(map[domain.RootName]domain.Path),
	}
}

func (p *RootResolver) ResolveRoot(rootName domain.RootName, relativePath domain.Path) (domain.Path, error) {
	rootPath, ok := p.rootMap[rootName]
	if !ok {
		return "", RootNotFoundError{
			RootName: rootName,
		}
	}

	rootPathWithRelative, err := rootPath.Join(relativePath.String())
	if err != nil {
		return "", err
	}

	return rootPathWithRelative, nil
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
