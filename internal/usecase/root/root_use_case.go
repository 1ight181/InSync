package root

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

func (r *RootUseCase) AddRoot(rootName string, rootPath string) {
	r.rootRegistrar.AddRoot(rootName, rootPath)
}

func (r *RootUseCase) RemoveRoot(rootName string) {
	r.rootRegistrar.RemoveRoot(rootName)
}
