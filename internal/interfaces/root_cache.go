package interfaces

type IRootCache interface {
	GetRootCache(rootName string) (string, error)
	SetRootCache(rootName string, path string) error
}
