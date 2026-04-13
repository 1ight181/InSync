package root

import "insync/internal/domain"

type RootUseCase struct {
	rootRegistrar IRootRegistrar
}

type RootUseCaseOptions struct {
	RootRegistrar IRootRegistrar
}

func NewRootUseCase(opts RootUseCaseOptions) *RootUseCase {
	if opts.RootRegistrar == nil {
		panic("Не все обязательные параметры были переданы при инициализации RootUseCase")
	}
	return &RootUseCase{rootRegistrar: opts.RootRegistrar}
}

func (r *RootUseCase) AddRoot(rootName domain.RootName, rootPath domain.Path) {
	r.rootRegistrar.AddRoot(rootName, rootPath)
}

func (r *RootUseCase) RemoveRoot(rootName domain.RootName) {
	r.rootRegistrar.RemoveRoot(rootName)
}

func (r *RootUseCase) GetRoots() []domain.RootName {
	return r.rootRegistrar.GetRoots()
}
