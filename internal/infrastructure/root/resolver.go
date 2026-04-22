package root

import (
	"insync/internal/domain"
)

type RootResolver struct {
	rootMap  map[domain.RootName]domain.Path
	rootRepo IRootResolverRepository
}

func NewRootResolver(rootRepo IRootResolverRepository) (*RootResolver, error) {
	rootResolver := RootResolver{
		rootMap:  make(map[domain.RootName]domain.Path),
		rootRepo: rootRepo,
	}

	roots, err := rootResolver.rootRepo.GetRoots()
	if err != nil {
		return nil, err
	}

	for rootName, rootPath := range roots {
		rootResolver.rootMap[rootName] = rootPath
	}

	return &rootResolver, nil
}

func (p *RootResolver) ResolveRoot(scopedPath domain.ScopedPath) (domain.Path, error) {
	rootPath, ok := p.rootMap[scopedPath.Root]
	if !ok {
		return "", RootNotFoundError{
			RootName: scopedPath.Root,
		}
	}

	rootPathWithRelative, err := rootPath.Join(scopedPath.Path.String())
	if err != nil {
		return "", err
	}

	return rootPathWithRelative, nil
}

func (p *RootResolver) AddRoot(rootName domain.RootName, rootPath domain.Path) {
	p.rootRepo.AddRoot(rootName, rootPath)
	p.rootMap[rootName] = rootPath
}

func (p *RootResolver) RemoveRoot(rootName domain.RootName) {
	p.rootRepo.RemoveRoot(rootName)
	delete(p.rootMap, rootName)
}

func (p *RootResolver) GetRoots() []domain.RootName {
	roots := make([]domain.RootName, 0, len(p.rootMap))
	for rootName := range p.rootMap {
		roots = append(roots, rootName)
	}
	return roots
}
