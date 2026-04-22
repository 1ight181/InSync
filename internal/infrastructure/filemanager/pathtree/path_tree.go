package pathtree

import (
	"insync/internal/domain"
	"path/filepath"
)

type ScopedPath struct {
	RootName domain.RootName
	Path     domain.Path
}

type node struct {
	path     ScopedPath
	parent   *node
	children map[ScopedPath]*node
}

type PathTree struct {
	nodes map[ScopedPath]*node
}

func NewPathTree() *PathTree {
	return &PathTree{
		nodes: make(map[ScopedPath]*node),
	}
}

func (pt *PathTree) AddPath(rootName domain.RootName, relativePath domain.Path) error {
	normalizedPath, err := pt.normalizePath(relativePath)
	if err != nil {
		return err
	}

	scoped := ScopedPath{
		RootName: rootName,
		Path:     normalizedPath,
	}

	if _, exists := pt.nodes[scoped]; exists {
		return nil
	}

	return pt.ensurePathRecursive(scoped)
}

func (pt *PathTree) RemovePath(rootName domain.RootName, relativePath domain.Path) error {
	normalizedPath, err := pt.normalizePath(relativePath)
	if err != nil {
		return err
	}

	scoped := ScopedPath{
		RootName: rootName,
		Path:     normalizedPath,
	}

	targetNode, exists := pt.nodes[scoped]
	if !exists {
		return ErrNotFound
	}

	pt.removeNodeRecursive(targetNode)
	return nil
}

func (pt *PathTree) GetParents(rootName domain.RootName, relativePath domain.Path) ([]ScopedPath, error) {
	normalizedPath, err := pt.normalizePath(relativePath)
	if err != nil {
		return nil, err
	}

	scoped := ScopedPath{
		RootName: rootName,
		Path:     normalizedPath,
	}

	currentNode := pt.nodes[scoped]
	if currentNode == nil {
		return nil, ErrNotFound
	}

	var parents []ScopedPath
	parentNode := currentNode.parent

	for parentNode != nil {
		parents = append(parents, parentNode.path)
		parentNode = parentNode.parent
	}

	return parents, nil
}

func (pt *PathTree) getChildren(rootName domain.RootName, relativePath domain.Path) ([]ScopedPath, error) {
	normalizedPath, err := pt.normalizePath(relativePath)
	if err != nil {
		return nil, err
	}

	scoped := ScopedPath{
		RootName: rootName,
		Path:     normalizedPath,
	}

	currentNode := pt.nodes[scoped]
	if currentNode == nil {
		return nil, ErrNotFound
	}

	children := make([]ScopedPath, 0, len(currentNode.children))
	for childPath := range currentNode.children {
		children = append(children, childPath)
	}

	return children, nil
}

func (pt *PathTree) ensurePathRecursive(current ScopedPath) error {
	if _, exists := pt.nodes[current]; exists {
		return nil
	}

	parentScoped, hasParent, err := pt.parentScopedPath(current)
	if err != nil {
		return err
	}

	if hasParent {
		if err := pt.ensurePathRecursive(parentScoped); err != nil {
			return err
		}
	}

	var parentNode *node
	if hasParent {
		parentNode = pt.nodes[parentScoped]
	}

	newNode := &node{
		path:     current,
		parent:   parentNode,
		children: make(map[ScopedPath]*node),
	}

	pt.nodes[current] = newNode

	if parentNode != nil {
		parentNode.children[current] = newNode
	}

	return nil
}

func (pt *PathTree) removeNodeRecursive(targetNode *node) {
	for _, childNode := range targetNode.children {
		pt.removeNodeRecursive(childNode)
	}

	if targetNode.parent != nil {
		delete(targetNode.parent.children, targetNode.path)
	}

	delete(pt.nodes, targetNode.path)
}

func (pt *PathTree) normalizePath(path domain.Path) (domain.Path, error) {
	if filepath.IsAbs(path.String()) {
		return "", ErrAbsPath
	}

	cleaned := path.Clean()
	if cleaned == "." {
		return "", nil
	}

	return cleaned, nil
}

func (pt *PathTree) parentScopedPath(current ScopedPath) (ScopedPath, bool, error) {
	parentPath := current.Path.Dir()

	if parentPath == "." || parentPath == "" {
		return ScopedPath{}, false, nil
	}

	return ScopedPath{
		RootName: current.RootName,
		Path:     parentPath,
	}, true, nil
}
