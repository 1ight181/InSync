package root

type IRootRegistrar interface {
	AddRoot(rootName string, rootPath string)
	RemoveRoot(rootName string)
}
