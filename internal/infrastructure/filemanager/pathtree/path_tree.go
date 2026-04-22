package pathtree

import (
	"insync/internal/domain"
)

type node struct {
	path     domain.ScopedPath
	parent   *node
	children map[domain.ScopedPath]*node
}

type PathTree struct {
	nodes map[domain.ScopedPath]*node
}

func NewPathTree() *PathTree {
	return &PathTree{
		nodes: make(map[domain.ScopedPath]*node),
	}
}

func (pt *PathTree) AddPath(scopedPath domain.ScopedPath) error {
	if _, exists := pt.nodes[scopedPath]; exists {
		return nil
	}

	return pt.ensurePathRecursive(scopedPath)
}

func (pt *PathTree) RemovePath(scopedPath domain.ScopedPath) error {
	targetNode, exists := pt.nodes[scopedPath]
	if !exists {
		return ErrNotFound
	}

	pt.removeNodeRecursive(targetNode)
	return nil
}

func (pt *PathTree) GetParents(scopedPath domain.ScopedPath) ([]domain.ScopedPath, error) {
	currentNode := pt.nodes[scopedPath]
	if currentNode == nil {
		return nil, ErrNotFound
	}

	var parents []domain.ScopedPath
	parentNode := currentNode.parent

	for parentNode != nil {
		parents = append(parents, parentNode.path)
		parentNode = parentNode.parent
	}

	return parents, nil
}

func (pt *PathTree) GetChildren(scopedPath domain.ScopedPath) ([]domain.ScopedPath, error) {
	currentNode := pt.nodes[scopedPath]
	if currentNode == nil {
		return nil, ErrNotFound
	}

	children := make([]domain.ScopedPath, 0, len(currentNode.children))
	for child := range currentNode.children {
		children = append(children, child)
	}

	return children, nil
}

func (pt *PathTree) ensurePathRecursive(current domain.ScopedPath) error {
	if _, exists := pt.nodes[current]; exists {
		return nil
	}

	parentScoped, hasParent := pt.parentScopedPath(current)

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
		children: make(map[domain.ScopedPath]*node),
	}

	pt.nodes[current] = newNode

	if parentNode != nil {
		parentNode.children[current] = newNode
	}

	return nil
}

func (pt *PathTree) removeNodeRecursive(targetNode *node) {
	for _, child := range targetNode.children {
		pt.removeNodeRecursive(child)
	}

	if targetNode.parent != nil {
		delete(targetNode.parent.children, targetNode.path)
	}

	delete(pt.nodes, targetNode.path)
}

func (pt *PathTree) parentScopedPath(current domain.ScopedPath) (domain.ScopedPath, bool) {
	parentPath := current.Path.Dir()

	if parentPath == "." || parentPath == "" {
		return domain.ScopedPath{}, false
	}

	scopedPath, err := domain.NewScopedPath(current.Root, parentPath)
	if err != nil {
		return domain.ScopedPath{}, false
	}
	return scopedPath, true
}
