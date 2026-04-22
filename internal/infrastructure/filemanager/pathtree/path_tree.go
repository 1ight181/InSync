package pathtree

import (
	"errors"
	"insync/internal/domain"
	"path/filepath"
)

type node struct {
	path     domain.Path
	parent   *node
	children map[domain.Path]*node
}

type PathTree struct {
	nodes map[domain.Path]*node
}

func NewPathTree() *PathTree {
	return &PathTree{
		nodes: make(map[domain.Path]*node),
	}
}

func (pt *PathTree) AddPath(fullPath domain.Path) error {
	normalizedPath, err := pt.normalizePath(fullPath)
	if err != nil {
		return err
	}
	if _, ok := pt.nodes[normalizedPath]; ok {
		return nil
	}

	pt.ensurePathRecursive(normalizedPath)
	return nil
}

func (pt *PathTree) RemovePath(fullPath domain.Path) error {
	normalizedPath, err := pt.normalizePath(fullPath)
	if err != nil {
		return err
	}
	node, exists := pt.nodes[normalizedPath]
	if !exists {
		return ErrNotFound
	}

	pt.removeNode(node)
	return nil
}

func (pt *PathTree) removeNode(node *node) {
	for _, childNode := range node.children {
		pt.removeNode(childNode)
	}

	if node.parent != nil {
		delete(node.parent.children, node.path)
	}
	delete(pt.nodes, node.path)
}

func (tree *PathTree) GetParents(fullPath domain.Path) ([]domain.Path, error) {
	normalizedPath, err := tree.normalizePath(fullPath)
	if err != nil {
		return nil, err
	}

	node := tree.nodes[normalizedPath]
	if node == nil {
		return nil, ErrNotFound
	}

	var parents []domain.Path
	current := node.parent

	for current != nil {
		parents = append(parents, current.path)
		current = current.parent
	}

	return parents, nil
}

func (tree *PathTree) getChildren(fullPath domain.Path) ([]domain.Path, error) {
	normalizedPath, err := tree.normalizePath(fullPath)
	if err != nil {
		return nil, err
	}
	node := tree.nodes[normalizedPath]
	if node == nil {
		return nil, ErrNotFound
	}

	var children []domain.Path
	for childPath := range node.children {
		children = append(children, childPath)
	}
	return children, nil
}

func (pt *PathTree) ensurePathRecursive(path domain.Path) error {
	if _, exists := pt.nodes[path]; exists {
		return nil
	}

	parent, err := pt.parentPath(path)
	if err != nil {
		if !errors.Is(err, ErrParentNotFound) {
			return err
		}
	}

	if errors.Is(err, ErrParentNotFound) {
		parentNode := pt.nodes[parent]
		newNode := &node{
			path:     path,
			parent:   parentNode,
			children: make(map[domain.Path]*node),
		}

		pt.nodes[path] = newNode
		if parentNode != nil {
			parentNode.children[path] = newNode
		}

		return nil
	}

	pt.ensurePathRecursive(parent)

	parentNode := pt.nodes[parent]
	newNode := &node{
		path:     path,
		parent:   parentNode,
		children: make(map[domain.Path]*node),
	}

	pt.nodes[path] = newNode
	if parentNode != nil {
		parentNode.children[path] = newNode
	}

	return nil

}

func (pt *PathTree) normalizePath(path domain.Path) (domain.Path, error) {
	if filepath.IsAbs(path.String()) {
		return "", ErrAbsPath
	}

	cleaned := path.Clean()

	return cleaned, nil
}

func (pt *PathTree) parentPath(fullPath domain.Path) (domain.Path, error) {
	cleaned, err := pt.normalizePath(fullPath)
	if err != nil {
		return "", err
	}

	parent := cleaned.Dir()

	if parent == "." {
		return "", ErrParentNotFound
	}

	return parent, nil
}
