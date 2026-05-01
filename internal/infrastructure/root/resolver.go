package root

import (
	"context"
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

func (p *RootResolver) AddRoot(ctx context.Context, rootName domain.RootName, rootPath domain.Path) error {
	p.rootMap[rootName] = rootPath
	return p.rootRepo.AddRoot(ctx, rootName, rootPath)
}

func (p *RootResolver) RemoveRoot(ctx context.Context, rootName domain.RootName) error {
	delete(p.rootMap, rootName)
	return p.rootRepo.RemoveRoot(ctx, rootName)
}

func (p *RootResolver) GetRoots() (map[domain.RootName]domain.Path, error) {
	if len(p.rootMap) == 0 {
		return nil, domain.ErrNoRoots
	}

	return p.rootMap, nil
}
