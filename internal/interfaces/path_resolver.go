package interfaces

type IPathResolver interface {
	ResolvePath(rootName string, relativePath string) (string, error)
}
