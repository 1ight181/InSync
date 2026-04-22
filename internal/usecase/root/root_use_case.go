package root

import (
	"errors"
	"insync/internal/domain"
)

type RootUseCase struct {
	rootRegistrar IRootRegistrar
}

type RootUseCaseOptions struct {
	RootRegistrar IRootRegistrar
}

var (
	ErrInvalidRootUseCaseOptions = errors.New("Все поля RootUseCase должны быть заполнены")
)

func NewRootUseCase(opts RootUseCaseOptions) (*RootUseCase, error) {
	if opts.RootRegistrar == nil {
		return nil, ErrInvalidRootUseCaseOptions
	}
	return &RootUseCase{rootRegistrar: opts.RootRegistrar}, nil
}

func (r *RootUseCase) AddRoot(rootName domain.RootName, rootPath domain.Path) error {
	return r.rootRegistrar.AddRoot(rootName, rootPath)
}

func (r *RootUseCase) RemoveRoot(rootName domain.RootName) error {
	return r.rootRegistrar.RemoveRoot(rootName)
}

func (r *RootUseCase) GetRoots() map[domain.RootName]domain.Path {
	return r.rootRegistrar.GetRoots()
}
