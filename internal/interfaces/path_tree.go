package interfaces

type IPathTreeReader interface {
	GetChildren(fullPath string) ([]string, error)
	GetParents(fullPath string) ([]string, error)
}

type IPathTreeWriter interface {
	AddPath(fullPath string) error
	RemovePath(fullPath string) error
}

type IPathTree interface {
	IPathTreeReader
	IPathTreeWriter
}
