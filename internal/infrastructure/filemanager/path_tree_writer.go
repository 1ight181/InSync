package filemanager

type IPathTreeWriter interface {
	AddPath(fullPath string) error
	RemovePath(fullPath string) error
}
