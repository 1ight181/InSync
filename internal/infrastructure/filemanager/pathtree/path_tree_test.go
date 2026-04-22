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

func mustScopedPath(t *testing.T, root domain.RootName, rawPath string) domain.ScopedPath {
	t.Helper()

	path := mustPath(t, rawPath)
	scopedPath, err := domain.NewScopedPath(root, path)
	require.NoError(t, err)
	return scopedPath
}

func TestPathTree_AddingNestedPath_CreatesFullParentChain(t *testing.T) {
	tree := NewPathTree()

	root := domain.RootName("root-a")
	leafPath := mustPath(t, filepath.Join("level1", "level2", "leaf.txt"))

	require.NoError(t, tree.AddPath(mustScopedPath(t, root, leafPath.String())))

	parents, err := tree.GetParents(mustScopedPath(t, root, leafPath.String()))
	require.NoError(t, err)

	expectedParents := []domain.ScopedPath{
		mustScopedPath(t, root, filepath.Join("level1", "level2")),
		mustScopedPath(t, root, filepath.Join("level1")),
	}

	assert.Equal(t, expectedParents, parents)

	parentDir := leafPath.Dir()

	children, err := tree.GetChildren(mustScopedPath(t, root, parentDir.String()))
	require.NoError(t, err)

	assert.ElementsMatch(t, []domain.ScopedPath{
		mustScopedPath(t, root, leafPath.String()),
	}, children)
}

func TestPathTree_AddingSamePathTwice_StateRemainsUnchanged(t *testing.T) {
	tree := NewPathTree()

	root := domain.RootName("root-a")
	leafPath := mustPath(t, filepath.Join("folder", "leaf.txt"))

	scoped := mustScopedPath(t, root, leafPath.String())

	require.NoError(t, tree.AddPath(scoped))

	parentsBefore, err := tree.GetParents(scoped)
	require.NoError(t, err)

	require.NoError(t, tree.AddPath(scoped))

	parentsAfter, err := tree.GetParents(scoped)
	require.NoError(t, err)

	assert.Equal(t, parentsBefore, parentsAfter)
}

func TestPathTree_PathIsNotNormalized_ItIsNormalizedToCanonicalForm(t *testing.T) {
	tree := NewPathTree()

	root := domain.RootName("root-a")

	cleanPath := mustPath(t, filepath.Join("folder", "leaf.txt"))
	messyPath := mustPath(t, filepath.Join("folder", "..", "folder", ".", "leaf.txt"))

	require.NoError(t, tree.AddPath(mustScopedPath(t, root, messyPath.String())))

	parents, err := tree.GetParents(mustScopedPath(t, root, cleanPath.String()))
	require.NoError(t, err)

	expected := []domain.ScopedPath{
		mustScopedPath(t, root, "folder"),
	}

	assert.Equal(t, expected, parents)
}

func TestPathTree_QueryingMissingPath_ReturnsNotFound(t *testing.T) {
	tree := NewPathTree()

	root := domain.RootName("root-a")
	missingPath := mustPath(t, filepath.Join("missing", "node.txt"))

	parents, err := tree.GetParents(mustScopedPath(t, root, missingPath.String()))
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, parents)

	children, err := tree.GetChildren(mustScopedPath(t, root, missingPath.String()))
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, children)
}

func TestPathTree_RemovingParentNode_RemovesEntireSubtree(t *testing.T) {
	tree := NewPathTree()

	root := domain.RootName("root-a")

	parent := mustPath(t, "parent")
	firstChild := mustPath(t, filepath.Join("parent", "first.txt"))
	secondChild := mustPath(t, filepath.Join("parent", "first.txt", "second.txt"))

	require.NoError(t, tree.AddPath(mustScopedPath(t, root, firstChild.String())))
	require.NoError(t, tree.AddPath(mustScopedPath(t, root, parent.String())))
	require.NoError(t, tree.AddPath(mustScopedPath(t, root, secondChild.String())))

	childrenBefore, err := tree.GetChildren(mustScopedPath(t, root, parent.String()))
	require.NoError(t, err)

	assert.ElementsMatch(t, []domain.ScopedPath{
		mustScopedPath(t, root, firstChild.String()),
	}, childrenBefore)

	require.NoError(t, tree.RemovePath(mustScopedPath(t, root, parent.String())))

	_, err = tree.GetParents(mustScopedPath(t, root, firstChild.String()))
	require.ErrorIs(t, err, ErrNotFound)

	_, err = tree.GetParents(mustScopedPath(t, root, secondChild.String()))
	require.ErrorIs(t, err, ErrNotFound)

	_, err = tree.GetChildren(mustScopedPath(t, root, parent.String()))
	require.ErrorIs(t, err, ErrNotFound)
}

func TestPathTree_RemovingUnknownPath_ReturnsNotFound(t *testing.T) {
	tree := NewPathTree()

	root := domain.RootName("root-a")
	missingPath := mustPath(t, filepath.Join("missing", "node.txt"))

	err := tree.RemovePath(mustScopedPath(t, root, missingPath.String()))
	require.ErrorIs(t, err, ErrNotFound)
}

func TestPathTree_SameRelativePathExistsInDifferentRoots_TheyAreIsolated(t *testing.T) {
	tree := NewPathTree()

	rootA := domain.RootName("root-a")
	rootB := domain.RootName("root-b")

	shared := mustPath(t, filepath.Join("folder", "leaf.txt"))

	require.NoError(t, tree.AddPath(mustScopedPath(t, rootA, shared.String())))
	require.NoError(t, tree.AddPath(mustScopedPath(t, rootB, shared.String())))

	parentsA, err := tree.GetParents(mustScopedPath(t, rootA, shared.String()))
	require.NoError(t, err)

	parentsB, err := tree.GetParents(mustScopedPath(t, rootB, shared.String()))
	require.NoError(t, err)

	assert.Equal(t, []domain.ScopedPath{
		mustScopedPath(t, rootA, "folder"),
	}, parentsA)

	assert.Equal(t, []domain.ScopedPath{
		mustScopedPath(t, rootB, "folder"),
	}, parentsB)
}

func TestPathTree_RemovingPathInOneRoot_OtherRootRemainsUnaffected(t *testing.T) {
	tree := NewPathTree()

	rootA := domain.RootName("root-a")
	rootB := domain.RootName("root-b")

	parent := mustPath(t, "folder")
	leaf := mustPath(t, filepath.Join("folder", "leaf.txt"))

	require.NoError(t, tree.AddPath(mustScopedPath(t, rootA, leaf.String())))
	require.NoError(t, tree.AddPath(mustScopedPath(t, rootB, leaf.String())))
	require.NoError(t, tree.AddPath(mustScopedPath(t, rootA, parent.String())))
	require.NoError(t, tree.AddPath(mustScopedPath(t, rootB, parent.String())))

	require.NoError(t, tree.RemovePath(mustScopedPath(t, rootA, parent.String())))

	_, err := tree.GetParents(mustScopedPath(t, rootA, leaf.String()))
	require.ErrorIs(t, err, ErrNotFound)

	parentsB, err := tree.GetParents(mustScopedPath(t, rootB, leaf.String()))
	require.NoError(t, err)

	assert.Equal(t, []domain.ScopedPath{
		mustScopedPath(t, rootB, "folder"),
	}, parentsB)
}
