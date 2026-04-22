package pathtree

import (
	"path/filepath"
	"testing"

	"insync/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustPath(t *testing.T, rawPath string) domain.Path {
	t.Helper()

	createdPath, err := domain.NewPath(rawPath)
	require.NoError(t, err)

	return createdPath
}

func TestPathTree_AddingNestedPath_CreatesFullParentChain(t *testing.T) {
	tree := NewPathTree()

	rootName := domain.RootName("root-a")
	leafPath := mustPath(t, filepath.Join("level1", "level2", "leaf.txt"))

	require.NoError(t, tree.AddPath(rootName, leafPath))

	parents, err := tree.GetParents(rootName, leafPath)
	require.NoError(t, err)

	expectedParents := []ScopedPath{
		{RootName: rootName, Path: mustPath(t, filepath.Join("level1", "level2"))},
		{RootName: rootName, Path: mustPath(t, filepath.Join("level1"))},
	}
	assert.Equal(t, expectedParents, parents)

	parentDir := leafPath.Dir()
	children, err := tree.getChildren(rootName, parentDir)
	require.NoError(t, err)

	assert.ElementsMatch(t, []ScopedPath{
		{RootName: rootName, Path: leafPath},
	}, children)
}

func TestPathTree_AddingSamePathTwice_StateRemainsUnchanged(t *testing.T) {
	tree := NewPathTree()

	rootName := domain.RootName("root-a")
	leafPath := mustPath(t, filepath.Join("folder", "leaf.txt"))

	require.NoError(t, tree.AddPath(rootName, leafPath))

	parentsBefore, err := tree.GetParents(rootName, leafPath)
	require.NoError(t, err)

	require.NoError(t, tree.AddPath(rootName, leafPath))

	parentsAfter, err := tree.GetParents(rootName, leafPath)
	require.NoError(t, err)

	assert.Equal(t, parentsBefore, parentsAfter)
}

func TestPathTree_PathIsNotNormalized_ItIsNormalizedToCanonicalForm(t *testing.T) {
	tree := NewPathTree()

	rootName := domain.RootName("root-a")
	cleanPath := mustPath(t, filepath.Join("folder", "leaf.txt"))
	messyPath := mustPath(t, filepath.Join("folder", "..", "folder", ".", "leaf.txt"))

	require.NoError(t, tree.AddPath(rootName, messyPath))

	parents, err := tree.GetParents(rootName, cleanPath)
	require.NoError(t, err)

	expected := []ScopedPath{
		{RootName: rootName, Path: mustPath(t, "folder")},
	}
	assert.Equal(t, expected, parents)
}

func TestPathTree_QueryingMissingPath_ReturnsNotFound(t *testing.T) {
	tree := NewPathTree()

	rootName := domain.RootName("root-a")
	missingPath := mustPath(t, filepath.Join("missing", "node.txt"))

	parents, err := tree.GetParents(rootName, missingPath)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, parents)

	children, err := tree.getChildren(rootName, missingPath)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, children)
}

func TestPathTree_RemovingParentNode_RemovesEntireSubtree(t *testing.T) {
	tree := NewPathTree()

	rootName := domain.RootName("root-a")

	parentPath := mustPath(t, "parent")
	firstChildPath := mustPath(t, filepath.Join("parent", "first.txt"))
	secondChildPath := mustPath(t, filepath.Join("parent", "first.txt", "second.txt"))

	require.NoError(t, tree.AddPath(rootName, firstChildPath))
	require.NoError(t, tree.AddPath(rootName, parentPath))
	require.NoError(t, tree.AddPath(rootName, secondChildPath))

	childrenBefore, err := tree.getChildren(rootName, parentPath)
	require.NoError(t, err)

	assert.ElementsMatch(t, []ScopedPath{
		{RootName: rootName, Path: firstChildPath},
	}, childrenBefore)

	require.NoError(t, tree.RemovePath(rootName, parentPath))

	_, err = tree.GetParents(rootName, firstChildPath)
	require.ErrorIs(t, err, ErrNotFound)

	_, err = tree.GetParents(rootName, secondChildPath)
	require.ErrorIs(t, err, ErrNotFound)

	_, err = tree.getChildren(rootName, parentPath)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestPathTree_RemovingUnknownPath_ReturnsNotFound(t *testing.T) {
	tree := NewPathTree()

	rootName := domain.RootName("root-a")
	missingPath := mustPath(t, filepath.Join("missing", "node.txt"))

	err := tree.RemovePath(rootName, missingPath)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestPathTree_AddingAbsolutePath_ReturnsError(t *testing.T) {
	tree := NewPathTree()

	rootName := domain.RootName("root-a")

	absPath, err := filepath.Abs("abs/path")
	require.NoError(t, err)
	require.NotEmpty(t, absPath)

	err = tree.AddPath(rootName, domain.Path(absPath))
	require.ErrorIs(t, err, ErrAbsPath)
}

func TestPathTree_SameRelativePathExistsInDifferentRoots_TheyAreIsolated(t *testing.T) {
	tree := NewPathTree()

	rootA := domain.RootName("root-a")
	rootB := domain.RootName("root-b")

	sharedPath := mustPath(t, filepath.Join("folder", "leaf.txt"))

	require.NoError(t, tree.AddPath(rootA, sharedPath))
	require.NoError(t, tree.AddPath(rootB, sharedPath))

	parentsA, err := tree.GetParents(rootA, sharedPath)
	require.NoError(t, err)

	parentsB, err := tree.GetParents(rootB, sharedPath)
	require.NoError(t, err)

	assert.Equal(t, []ScopedPath{
		{RootName: rootA, Path: mustPath(t, "folder")},
	}, parentsA)

	assert.Equal(t, []ScopedPath{
		{RootName: rootB, Path: mustPath(t, "folder")},
	}, parentsB)

	childrenA, err := tree.getChildren(rootA, mustPath(t, "folder"))
	require.NoError(t, err)
	assert.ElementsMatch(t, []ScopedPath{
		{RootName: rootA, Path: sharedPath},
	}, childrenA)

	childrenB, err := tree.getChildren(rootB, mustPath(t, "folder"))
	require.NoError(t, err)
	assert.ElementsMatch(t, []ScopedPath{
		{RootName: rootB, Path: sharedPath},
	}, childrenB)
}

func TestPathTree_RemovingPathInOneRoot_OtherRootRemainsUnaffected(t *testing.T) {
	tree := NewPathTree()

	rootA := domain.RootName("root-a")
	rootB := domain.RootName("root-b")

	parentPath := mustPath(t, "folder")
	leafPath := mustPath(t, filepath.Join("folder", "leaf.txt"))

	require.NoError(t, tree.AddPath(rootA, leafPath))
	require.NoError(t, tree.AddPath(rootB, leafPath))
	require.NoError(t, tree.AddPath(rootA, parentPath))
	require.NoError(t, tree.AddPath(rootB, parentPath))

	require.NoError(t, tree.RemovePath(rootA, parentPath))

	_, err := tree.GetParents(rootA, leafPath)
	require.ErrorIs(t, err, ErrNotFound)

	parentsB, err := tree.GetParents(rootB, leafPath)
	require.NoError(t, err)

	assert.Equal(t, []ScopedPath{
		{RootName: rootB, Path: mustPath(t, "folder")},
	}, parentsB)
}
