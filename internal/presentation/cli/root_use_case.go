package cli

type IRootUseCase interface {
	AddRoot(rootName string, rootPath string)
	RemoveRoot(rootName string)
}
