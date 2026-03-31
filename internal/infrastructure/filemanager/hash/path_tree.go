package hash

import (
	"insync/internal/interfaces"
	"path/filepath"
)

type node struct {
	path     string
	parent   *node
	children map[string]*node
}

type PathTree struct {
	nodes map[string]*node
}

func NewPathTree() interfaces.IPathTree {
	return &PathTree{
		nodes: make(map[string]*node),
	}
}

func (pt *PathTree) AddPath(fullPath string) error {
	normalizedPath, err := pt.normalizePath(fullPath)
	if err != nil {
		return err
	}
	if _, ok := pt.nodes[normalizedPath]; ok {
		return ErrAlreadyExists
	}

	pt.ensurePathRecursive(normalizedPath)
	return nil
}

func (pt *PathTree) RemovePath(fullPath string) error {
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

func (tree *PathTree) GetParents(fullPath string) ([]string, error) {
	normalizedPath, err := tree.normalizePath(fullPath)
	if err != nil {
		return nil, err
	}
	node := tree.nodes[normalizedPath]
	if node == nil {
		return nil, ErrNotFound
	}

	var parents []string
	current := node.parent

	for current != nil {
		parents = append(parents, current.path)
		current = current.parent
	}

	return parents, nil
}

func (tree *PathTree) GetChildren(fullPath string) ([]string, error) {
	normalizedPath, err := tree.normalizePath(fullPath)
	if err != nil {
		return nil, err
	}
	node := tree.nodes[normalizedPath]
	if node == nil {
		return nil, ErrNotFound
	}

	var children []string
	for childPath := range node.children {
		children = append(children, childPath)
	}
	return children, nil
}

func (pt *PathTree) ensurePathRecursive(path string) error {
	if _, exists := pt.nodes[path]; exists {
		return nil
	}

	parent, err := pt.parentPath(path)
	if err != nil {
		return err
	}

	if parent != "" {
		pt.ensurePathRecursive(parent)
	}

	parentNode := pt.nodes[parent]
	newNode := &node{
		path:     path,
		parent:   parentNode,
		children: make(map[string]*node),
	}
	pt.nodes[path] = newNode
	if parentNode != nil {
		parentNode.children[path] = newNode
	}

	return nil
}

func (pt *PathTree) normalizePath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	cleaned := filepath.Clean(absPath)

	if cleaned == "." {
		return "", nil
	}

	return cleaned, nil
}

func (pt *PathTree) parentPath(fullPath string) (string, error) {
	cleaned, err := pt.normalizePath(fullPath)
	if err != nil {
		return "", err
	}
	if cleaned == "" {
		return "", nil
	}

	parent := filepath.Dir(cleaned)

	if parent == "." || parent == cleaned || cleaned == "/" {
		return "", nil
	}

	return parent, nil
}
