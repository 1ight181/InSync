package root

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type RootUseCase struct {
	rootRegistrar IRootRegistrar
}

var (
	ErrInvalidRootUseCaseOptions = errors.New("Все поля RootUseCase должны быть заполнены")
)

func NewRootUseCase(rootRegistrar IRootRegistrar) (*RootUseCase, error) {
	if rootRegistrar == nil {
		return nil, ErrInvalidRootUseCaseOptions
	}
	return &RootUseCase{rootRegistrar: rootRegistrar}, nil
}

func (r *RootUseCase) AddRoot(ctx context.Context, rootName domain.RootName, rootPath domain.Path) error {
	return r.rootRegistrar.AddRoot(ctx, rootName, rootPath)
}

func (r *RootUseCase) RemoveRoot(ctx context.Context, rootName domain.RootName) error {
	return r.rootRegistrar.RemoveRoot(ctx, rootName)
}

func (r *RootUseCase) GetRoots() (map[domain.RootName]domain.Path, error) {
	return r.rootRegistrar.GetRoots()
}
