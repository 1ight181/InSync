package filemanager

type IRootResolver interface {
	ResolveRoot(rootName string, relativePath string) (string, error)
	AddRoot(rootName string, rootPath string)
	RemoveRoot(rootName string)
}
