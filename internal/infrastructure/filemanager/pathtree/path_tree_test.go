package pathtree

import (
	"path/filepath"
	"testing"

	"insync/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathTree_AddPath_CreatesNodeAndParents(t *testing.T) {
	tree := NewPathTree()

	leaf := domain.Path(filepath.Join("level1", "level2", "leaf.txt"))

	require.NoError(t, tree.AddPath(leaf))

	parents, err := tree.GetParents(leaf)
	require.NoError(t, err)

	// Проверяем цепочку родителей (от ближайшего к корню)
	expectedParents := []domain.Path{
		domain.Path(filepath.Join("level1", "level2")),
		domain.Path(filepath.Join("level1")),
	}
	assert.Equal(t, expectedParents, parents)

	// Проверяем детей у родителя
	parentDir := leaf.Dir()
	children, err := tree.getChildren(parentDir)
	require.NoError(t, err)
	assert.ElementsMatch(t, []domain.Path{leaf}, children)
}

func TestPathTree_AddPath_IsIdempotent(t *testing.T) {
	tree := NewPathTree()

	leaf := domain.Path(filepath.Join("folder", "leaf.txt"))

	require.NoError(t, tree.AddPath(leaf))

	parentsBefore, err := tree.GetParents(leaf)
	require.NoError(t, err)

	require.NoError(t, tree.AddPath(leaf))

	parentsAfter, err := tree.GetParents(leaf)
	require.NoError(t, err)

	assert.Equal(t, parentsBefore, parentsAfter)
}

func TestPathTree_AddPath_NormalizesInputPath(t *testing.T) {
	tree := NewPathTree()

	cleanRaw := domain.Path(filepath.Join("folder", "leaf.txt"))
	messyRaw := domain.Path(filepath.Join("folder", "..", "folder", ".", "leaf.txt"))

	require.NoError(t, tree.AddPath(messyRaw))

	// Должно нормализоваться к одному и тому же пути
	parents, err := tree.GetParents(cleanRaw)
	require.NoError(t, err)

	expected := []domain.Path{
		domain.Path("folder"),
	}
	assert.Equal(t, expected, parents)
}

func TestPathTree_GetParentsAndgetChildren_NotFound(t *testing.T) {
	tree := NewPathTree()

	missing := domain.Path(filepath.Join("missing", "node.txt"))

	parents, err := tree.GetParents(missing)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, parents)

	children, err := tree.getChildren(missing)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, children)
}

func TestPathTree_RemovePath_RemovesSubtree(t *testing.T) {
	tree := NewPathTree()

	parentRaw := domain.Path(filepath.Join("parent"))
	firstRaw, err := parentRaw.Join("first.txt")
	require.NoError(t, err)
	secondRaw, err := firstRaw.Join("second.txt")
	require.NoError(t, err)

	require.NoError(t, tree.AddPath(firstRaw))
	require.NoError(t, tree.AddPath(parentRaw))
	require.NoError(t, tree.AddPath(secondRaw))

	childrenBefore, err := tree.getChildren(parentRaw)
	require.NoError(t, err)
	assert.ElementsMatch(t, []domain.Path{firstRaw}, childrenBefore)

	require.NoError(t, tree.RemovePath(domain.Path(parentRaw)))

	_, err = tree.GetParents(domain.Path(firstRaw))
	require.ErrorIs(t, err, ErrNotFound)

	_, err = tree.GetParents(domain.Path(secondRaw))
	require.ErrorIs(t, err, ErrNotFound)

	_, err = tree.getChildren(domain.Path(parentRaw))
	require.ErrorIs(t, err, ErrNotFound)
}

func TestPathTree_RemovePath_UnknownPathReturnsErrNotFound(t *testing.T) {
	tree := NewPathTree()

	missing := domain.Path(filepath.Join("missing", "node.txt"))

	err := tree.RemovePath(missing)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestPathTree_AddAbsPath_Error(t *testing.T) {
	tree := NewPathTree()

	absPath, err := filepath.Abs("abs/path")
	require.NoError(t, err)
	require.NotEmpty(t, absPath)

	err = tree.AddPath(domain.Path(absPath))
	require.ErrorIs(t, err, ErrAbsPath)
}
