package scan

import (
	"context"
	"insync/internal/domain"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlanner_EmptySnapshots_EmptyPlan(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	emptyBaseSnap := createSnapshot(t, nil)
	emptyLocalSnap := createSnapshot(t, nil)
	emptyRemoteSnap := createSnapshot(t, nil)

	plan, err := planner.Plan(ctx, emptyBaseSnap, emptyLocalSnap, emptyRemoteSnap)
	require.NoError(t, err)
	require.Empty(t, plan)
}

func TestPlanner_SameSnapshots_EmptyPlan(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	sameFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 4, nil),
		createFileEntry(t, "/a/b", "hash2", 3, nil),
		createFileEntry(t, "/a/b/c", "hash3", 1, nil),
		createFileEntry(t, "/a/b/d", "hash4", 1, nil),
	}

	baseSnap := createSnapshot(t, sameFiles)
	localSnap := createSnapshot(t, sameFiles)
	remoteSnap := createSnapshot(t, sameFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)
	require.Empty(t, plan)
}

func TestPlanner_NewFileOnLocal_CreateChangeOnRemote(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	localFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/c", "hash3", 1, nil),
	}

	remoteFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)
	require.Len(t, plan.LocalChanges, 0)
	require.Len(t, plan.RemoteChanges, 1)
	require.Len(t, plan.Conflicts, 0)

	require.Equal(t, domain.Create, plan.RemoteChanges[0].ChangeType)
}

func TestPlanner_NewFileOnRemote_CreateChangeOnLocal(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	localFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	remoteFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/c", "hash3", 1, nil),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)
	require.Len(t, plan.LocalChanges, 1)
	require.Len(t, plan.RemoteChanges, 0)
	require.Len(t, plan.Conflicts, 0)

	require.Equal(t, domain.Create, plan.LocalChanges[0].ChangeType)
}

func TestPlanner_NewSameFileOnRemoteAndLocal_EmptyPlan(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	localFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/c", "hash3", 1, nil),
	}

	remoteFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/c", "hash3", 1, nil),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)
	require.Empty(t, plan)
}

func TestPlanner_NewFileWithDifferentContentButSameNameOnRemoteAndLocal_CreateChangeOmBothSides(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	localFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/c", "hash3", 1, nil),
	}

	remoteFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/c", "hash4", 1, nil),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)
	require.Len(t, plan.LocalChanges, 0)
	require.Len(t, plan.RemoteChanges, 0)
	require.Len(t, plan.Conflicts, 1)

	require.Equal(t, domain.ConflictBothCreatedAtSamePathConflict, plan.Conflicts[0].Conflict)
}

func TestPlanner_NewFileWithDifferentNameButSameContentOnRemoteAndLocal_CreateChangeOmBothSides(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	localFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/c", "hash3", 1, nil),
	}

	remoteFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/d", "hash3", 1, nil),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)
	require.Len(t, plan.LocalChanges, 1)
	require.Len(t, plan.RemoteChanges, 1)

	require.Equal(t, domain.Create, plan.LocalChanges[0].ChangeType)
	require.Equal(t, domain.Create, plan.RemoteChanges[0].ChangeType)
}

func TestPlanner_NewFileWithDifferentNameAndModButSameContentOnRemoteAndLocal_CreateChangeOmBothSides(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	localMetadata := createMetadata(10, 10, false)
	localFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/c", "hash3", 1, &localMetadata),
	}

	remoteMetadata := createMetadata(5, 10, false)
	remoteFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/d", "hash3", 1, &remoteMetadata),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)
	require.Len(t, plan.LocalChanges, 1)
	require.Len(t, plan.RemoteChanges, 1)

	require.Equal(t, domain.Create, plan.LocalChanges[0].ChangeType)
	require.Equal(t, domain.Create, plan.RemoteChanges[0].ChangeType)
}

func TestPlanner_ThreeNewFilesAndDirOnLocalOneFileOnRemote_CreateChangeOnRemote(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	dirMetadata := createMetadata(10, 10, true)

	localFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/c", "hash3", 1, nil),
		createFileEntry(t, "/a/b/d", "hash4", 1, nil),
		createFileEntry(t, "/a/b/e", "hash5", 1, &dirMetadata),
	}

	remoteFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, nil),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
		createFileEntry(t, "/a/b/f", "hash5", 1, nil),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 1)
	require.Len(t, plan.RemoteChanges, 3)

	expectedLocalChanges := domain.LocalChange{
		NewRelativePath: mustPath(t, "/a/b/f"),
		ChangeType:      domain.Create,
	}

	expectedRemoteChanges := []domain.RemoteChange{
		{
			NewRelativePath: mustPath(t, "/a/b/c"),
			ChangeType:      domain.Create,
		},
		{
			NewRelativePath: mustPath(t, "/a/b/d"),
			ChangeType:      domain.Create,
		},
		{
			NewRelativePath: mustPath(t, "/a/b/e"),
			ChangeType:      domain.Create,
		},
	}

	require.Equal(t, expectedLocalChanges, plan.LocalChanges[0])
	require.Equal(t, expectedRemoteChanges, plan.RemoteChanges)
}

// Занимательный результат: для папки отдает конфликт, что хэш разный, mtime одинаковый
// В реальной системе mtime не одинаковый, однако это верно только при
// добавлении/ренейму/делиту файлов внутри, но не изменению контента
// В целом это правильное поведение
// Но можно оптимизировать, проверяя размер папки

func TestPlanner_DeleteFileOnLocal_DeleteChangeOnRemote(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseMetadata := createMetadata(10, 10, true)
	baseFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, &baseMetadata),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	localMetadata := createMetadata(10, 10, true)
	localFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash3", 1, &localMetadata),
	}

	remoteMetadata := createMetadata(10, 10, true)
	remoteFiles := []domain.FileEntry{
		createFileEntry(t, "/a", "hash1", 2, &remoteMetadata),
		createFileEntry(t, "/a/b", "hash2", 1, nil),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 0)
	require.Len(t, plan.RemoteChanges, 1)
	require.Len(t, plan.Conflicts, 0)

	expectedRemoteChanges := domain.RemoteChange{
		OldRelativePath: mustPath(t, "/a/b"),
		ChangeType:      domain.Delete,
	}

	require.Equal(t, expectedRemoteChanges, plan.RemoteChanges[0])
}

// Возвращает пустой снапшот, если не заданы входные параметры
func createSnapshot(t *testing.T, fileEntries []domain.FileEntry) domain.Snapshot {
	t.Helper()
	if len(fileEntries) > 0 {
		return domain.Snapshot{
			Files: fileEntries,
		}
	}

	return domain.Snapshot{}
}

func createFileEntry(t *testing.T, relativePath string, hash string, subtreeSize uint64, metadata *domain.FileMetadata) domain.FileEntry {
	t.Helper()
	if metadata != nil {
		return domain.FileEntry{
			RelativePath: mustPath(t, relativePath),
			SubtreeSize:  subtreeSize,
			FileInfo: domain.FileInfo{
				Hash:     hash,
				Metadata: *metadata,
			},
		}
	}

	return domain.FileEntry{
		RelativePath: mustPath(t, relativePath),
		SubtreeSize:  subtreeSize,
		FileInfo: domain.FileInfo{
			Hash:     hash,
			Metadata: createMetadata(1, 1, false),
		},
	}
}

func mustPath(t *testing.T, rawPath string) domain.Path {
	t.Helper()
	p, err := domain.NewPath(rawPath)
	require.NoError(t, err)
	return p
}

func createMetadata(modUnix uint64, sizeBytes uint64, isDir bool) domain.FileMetadata {
	return domain.FileMetadata{
		ModifiedUnix: modUnix,
		SizeBytes:    sizeBytes,
		IsDirectory:  isDir,
	}
}
