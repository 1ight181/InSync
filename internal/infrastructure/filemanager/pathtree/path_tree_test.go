package pathtree

import (
	"path/filepath"
	"testing"

	"insync/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func normalize(t *testing.T, p string) domain.Path {
	t.Helper()
	dp := domain.Path(p)
	abs, err := dp.Abs()
	require.NoError(t, err)
	return abs
}

func TestPathTree_AddPath_CreatesNodeAndParents(t *testing.T) {
	tree := NewPathTree()

	base := t.TempDir()
	leafRaw := filepath.Join(base, "level1", "level2", "leaf.txt")

	require.NoError(t, tree.AddPath(domain.Path(leafRaw)))

	leaf := normalize(t, leafRaw)

	parents, err := tree.GetParents(domain.Path(leafRaw))
	require.NoError(t, err)

	// Проверяем цепочку родителей (от ближайшего к корню)
	expectedParents := []domain.Path{
		normalize(t, filepath.Join(base, "level1", "level2")),
		normalize(t, filepath.Join(base, "level1")),
		normalize(t, base),
	}
	assert.Equal(t, expectedParents, parents)

	// Проверяем детей у родителя
	parentDir := normalize(t, filepath.Dir(leafRaw))
	children, err := tree.GetChildren(parentDir)
	require.NoError(t, err)
	assert.ElementsMatch(t, []domain.Path{leaf}, children)
}

func TestPathTree_AddPath_IsIdempotent(t *testing.T) {
	tree := NewPathTree()

	base := t.TempDir()
	leafRaw := filepath.Join(base, "folder", "leaf.txt")

	leaf := normalize(t, leafRaw)

	require.NoError(t, tree.AddPath(domain.Path(leafRaw)))

	parentsBefore, err := tree.GetParents(domain.Path(leafRaw))
	require.NoError(t, err)

	// Добавляем повторно
	require.NoError(t, tree.AddPath(domain.Path(leafRaw)))

	parentsAfter, err := tree.GetParents(leaf)
	require.NoError(t, err)

	assert.Equal(t, parentsBefore, parentsAfter)
}

func TestPathTree_AddPath_NormalizesInputPath(t *testing.T) {
	tree := NewPathTree()

	base := t.TempDir()
	cleanRaw := filepath.Join(base, "folder", "leaf.txt")
	messyRaw := filepath.Join(base, "folder", "..", "folder", ".", "leaf.txt")

	require.NoError(t, tree.AddPath(domain.Path(messyRaw)))

	// Должно нормализоваться к одному и тому же пути
	parents, err := tree.GetParents(domain.Path(cleanRaw))
	require.NoError(t, err)

	expected := []domain.Path{
		normalize(t, filepath.Join(base, "folder")),
		normalize(t, base),
	}
	assert.Equal(t, expected, parents)
}

func TestPathTree_GetParentsAndGetChildren_NotFound(t *testing.T) {
	tree := NewPathTree()

	missing := domain.Path(filepath.Join(t.TempDir(), "missing", "node.txt"))

	parents, err := tree.GetParents(missing)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, parents)

	children, err := tree.GetChildren(missing)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, children)
}

func TestPathTree_RemovePath_RemovesSubtree(t *testing.T) {
	tree := NewPathTree()

	base := t.TempDir()
	parentRaw := filepath.Join(base, "parent")
	firstRaw := filepath.Join(parentRaw, "first.txt")
	secondRaw := filepath.Join(parentRaw, "second.txt")

	require.NoError(t, tree.AddPath(domain.Path(firstRaw)))
	require.NoError(t, tree.AddPath(domain.Path(secondRaw)))

	// Проверяем, что parent появился как ребёнок base
	basePath := normalize(t, base)
	childrenBefore, err := tree.GetChildren(basePath)
	require.NoError(t, err)
	assert.ElementsMatch(t, []domain.Path{normalize(t, parentRaw)}, childrenBefore)

	// Удаляем parent и всё поддерево
	require.NoError(t, tree.RemovePath(domain.Path(parentRaw)))

	// Всё поддерево должно исчезнуть
	_, err = tree.GetParents(domain.Path(firstRaw))
	require.ErrorIs(t, err, ErrNotFound)

	_, err = tree.GetParents(domain.Path(secondRaw))
	require.ErrorIs(t, err, ErrNotFound)

	_, err = tree.GetChildren(domain.Path(parentRaw))
	require.ErrorIs(t, err, ErrNotFound)

	// У base больше не должно быть детей
	childrenAfter, err := tree.GetChildren(basePath)
	require.NoError(t, err)
	assert.Empty(t, childrenAfter)
}

func TestPathTree_RemovePath_UnknownPathReturnsErrNotFound(t *testing.T) {
	tree := NewPathTree()

	missing := domain.Path(filepath.Join(t.TempDir(), "missing", "node.txt"))

	err := tree.RemovePath(missing)
	require.ErrorIs(t, err, ErrNotFound)
}
