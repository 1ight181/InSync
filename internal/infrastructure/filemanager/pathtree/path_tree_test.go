package pathtree

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func absCleanPath(t *testing.T, rawPath string) string {
	t.Helper()

	absolutePath, err := filepath.Abs(rawPath)
	require.NoError(t, err)

	return filepath.Clean(absolutePath)
}

func expectedParentChain(normalizedPath string) []string {
	var parents []string

	currentPath := filepath.Clean(normalizedPath)
	for {
		parentPath := filepath.Dir(currentPath)
		if parentPath == "." || parentPath == currentPath {
			break
		}

		parents = append(parents, parentPath)

		if parentPath == "/" {
			break
		}
		currentPath = parentPath
	}

	return parents
}

func TestPathTree_AddPath_CreatesNodeAndParents(t *testing.T) {
	tree := NewPathTree()

	baseDirectory := t.TempDir()
	leafPath := filepath.Join(baseDirectory, "level1", "level2", "leaf.txt")
	normalizedLeafPath := absCleanPath(t, leafPath)

	require.NoError(t, tree.AddPath(leafPath))

	parents, err := tree.GetParents(leafPath)
	require.NoError(t, err)
	assert.Equal(t, expectedParentChain(normalizedLeafPath), parents)

	children, err := tree.GetChildren(filepath.Dir(normalizedLeafPath))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{normalizedLeafPath}, children)
}

func TestPathTree_AddPath_IsIdempotent(t *testing.T) {
	tree := NewPathTree()

	baseDirectory := t.TempDir()
	leafPath := filepath.Join(baseDirectory, "folder", "leaf.txt")
	normalizedLeafPath := absCleanPath(t, leafPath)

	require.NoError(t, tree.AddPath(leafPath))
	parentsBefore, err := tree.GetParents(leafPath)
	require.NoError(t, err)

	require.NoError(t, tree.AddPath(leafPath))
	parentsAfter, err := tree.GetParents(normalizedLeafPath)
	require.NoError(t, err)

	assert.Equal(t, parentsBefore, parentsAfter)

	children, err := tree.GetChildren(filepath.Dir(normalizedLeafPath))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{normalizedLeafPath}, children)
}

func TestPathTree_AddPath_NormalizesInputPath(t *testing.T) {
	tree := NewPathTree()

	baseDirectory := t.TempDir()
	cleanPath := filepath.Join(baseDirectory, "folder", "leaf.txt")
	messyPath := filepath.Join(baseDirectory, "folder", "..", "folder", ".", "leaf.txt")

	normalizedCleanPath := absCleanPath(t, cleanPath)
	normalizedMessyPath := absCleanPath(t, messyPath)
	require.Equal(t, normalizedCleanPath, normalizedMessyPath)

	require.NoError(t, tree.AddPath(messyPath))

	parents, err := tree.GetParents(cleanPath)
	require.NoError(t, err)
	assert.Equal(t, expectedParentChain(normalizedCleanPath), parents)
}

func TestPathTree_GetParentsAndGetChildren_NotFound(t *testing.T) {
	tree := NewPathTree()

	missingPath := filepath.Join(t.TempDir(), "missing", "node.txt")

	parents, err := tree.GetParents(missingPath)
	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, parents)

	children, err := tree.GetChildren(missingPath)
	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, children)
}

func TestPathTree_RemovePath_RemovesSubtree(t *testing.T) {
	tree := NewPathTree()

	baseDirectory := t.TempDir()
	parentPath := filepath.Join(baseDirectory, "parent")
	firstChildPath := filepath.Join(parentPath, "first.txt")
	secondChildPath := filepath.Join(parentPath, "second.txt")

	require.NoError(t, tree.AddPath(firstChildPath))
	require.NoError(t, tree.AddPath(secondChildPath))

	baseChildrenBeforeRemoval, err := tree.GetChildren(baseDirectory)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{absCleanPath(t, parentPath)}, baseChildrenBeforeRemoval)

	require.NoError(t, tree.RemovePath(parentPath))

	_, err = tree.GetParents(firstChildPath)
	require.ErrorIs(t, err, ErrNotFound)

	_, err = tree.GetParents(secondChildPath)
	require.ErrorIs(t, err, ErrNotFound)

	_, err = tree.GetChildren(parentPath)
	require.ErrorIs(t, err, ErrNotFound)

	baseChildrenAfterRemoval, err := tree.GetChildren(baseDirectory)
	require.NoError(t, err)
	assert.Empty(t, baseChildrenAfterRemoval)
}

func TestPathTree_RemovePath_UnknownPathReturnsErrNotFound(t *testing.T) {
	tree := NewPathTree()

	missingPath := filepath.Join(t.TempDir(), "missing", "node.txt")

	err := tree.RemovePath(missingPath)
	require.ErrorIs(t, err, ErrNotFound)
}
