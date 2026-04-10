package hash

type IPathTreeReader interface {
	GetChildren(fullPath string) ([]string, error)
	GetParents(fullPath string) ([]string, error)
}
