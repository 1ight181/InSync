package root

import (
	interfaces "insync/internal/interfaces"
	"path/filepath"
)

type RootResolver struct {
	rootMap map[string]string
}

func NewRootResolver() interfaces.IRootResolver {
	return &RootResolver{
		rootMap: make(map[string]string),
	}
}

func (p *RootResolver) ResolveRoot(rootName string, relativePath string) (string, error) {
	rootPath, ok := p.rootMap[rootName]
	if !ok {
		return "", RootNotFoundError{
			RootName: rootName,
		}
	}

	return filepath.Join(rootPath, relativePath), nil
}

func (p *RootResolver) AddRoot(rootName string, rootPath string) {
	p.rootMap[rootName] = rootPath
}

func (p *RootResolver) RemoveRoot(rootName string) {
	delete(p.rootMap, rootName)
}
