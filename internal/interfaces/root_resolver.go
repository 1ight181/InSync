package interfaces

type IRootResolver interface {
	ResolveRoot(rootName string, relativePath string) (string, error)
}
