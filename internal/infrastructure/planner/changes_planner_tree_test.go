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
		createDirectoryEntry(t, "/a", "dir-a-v1", 4, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-v1", 3, 110, 0),
		createRegularFileEntry(t, "/a/b/c", "file-c-v1", 1, 120, 13),
		createRegularFileEntry(t, "/a/b/d", "file-d-v1", 1, 130, 17),
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
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-local", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-local", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/c", "file-c-v1", 1, 120, 31),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
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
	require.Equal(t, mustPath(t, "/a/b/c"), plan.RemoteChanges[0].NewRelativePath)
}

func TestPlanner_NewFileOnRemote_CreateChangeOnLocal(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-remote", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-remote", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/c", "file-c-v1", 1, 120, 31),
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
	require.Equal(t, mustPath(t, "/a/b/c"), plan.LocalChanges[0].NewRelativePath)
}

func TestPlanner_NewSameFileOnRemoteAndLocal_EmptyPlan(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-local", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-local", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/c", "file-c-v1", 1, 120, 31),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-remote", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-remote", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/c", "file-c-v1", 1, 120, 31),
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
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-local", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-local", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/c", "file-c-local", 1, 120, 31),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-remote", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-remote", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/c", "file-c-remote", 1, 121, 31),
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
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-local", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-local", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/c", "file-c-v1", 1, 120, 31),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-remote", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-remote", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/d", "file-c-v1", 1, 121, 31),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)
	require.Len(t, plan.LocalChanges, 1)
	require.Len(t, plan.RemoteChanges, 1)

	require.Equal(t, domain.Create, plan.LocalChanges[0].ChangeType)
	require.Equal(t, mustPath(t, "/a/b/d"), plan.LocalChanges[0].NewRelativePath)
	require.Equal(t, domain.Create, plan.RemoteChanges[0].ChangeType)
	require.Equal(t, mustPath(t, "/a/b/c"), plan.RemoteChanges[0].NewRelativePath)
}

func TestPlanner_NewFileWithDifferentNameAndModButSameContentOnRemoteAndLocal_CreateChangeOmBothSides(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-local", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-local", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/c", "file-c-v1", 1, 200, 31),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-remote", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-remote", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/d", "file-c-v1", 1, 150, 31),
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
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-local", 5, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-local", 4, 110, 0),
		createRegularFileEntry(t, "/a/b/c", "file-c-v1", 1, 120, 31),
		createRegularFileEntry(t, "/a/b/d", "file-d-v1", 1, 121, 32),
		createDirectoryEntry(t, "/a/b/e", "dir-e-local", 1, 130, 0),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-remote", 3, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-remote", 2, 110, 0),
		createRegularFileEntry(t, "/a/b/f", "file-f-v1", 1, 140, 33),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 1)
	require.Len(t, plan.RemoteChanges, 3)

	expectedLocalChanges := []domain.LocalChange{
		{
			NewRelativePath: mustPath(t, "/a/b/f"),
			ChangeType:      domain.Create,
		},
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

	require.ElementsMatch(t, expectedLocalChanges, plan.LocalChanges)
	require.ElementsMatch(t, expectedRemoteChanges, plan.RemoteChanges)
}

// Занимательный результат: для папки отдает конфликт, что хэш разный, mtime одинаковый
// В реальной системе mtime не одинаковый, однако это верно только при
// добавлении/ренейму/делиту файлов внутри, но не изменению контента
// В целом это правильное поведение
// Но можно оптимизировать, проверяя размер папки

func TestPlanner_DeleteFileOnLocal_DeleteChangeOnRemote(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-local-after-delete", 1, 200, 0),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
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

func TestPlanner_DeleteFileOnRemoteAndLocalNewFileOnLocalWithOldNameDifferentHash_Conflict(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-base", 2, 100, 0),
		createDirectoryEntry(t, "/a/b", "dir-a-b-base", 1, 110, 0),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-local-after-change", 2, 200, 0),
		createRegularFileEntry(t, "/a/b", "file-b-local", 1, 201, 44),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/a", "dir-a-base", 1, 100, 0),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 0)
	require.Len(t, plan.RemoteChanges, 0)
	require.Len(t, plan.Conflicts, 1)

	expectedConflict := domain.Conflict{
		BaseRelativePath:  mustPath(t, "/a/b"),
		LocalRelativePath: mustPath(t, "/a/b"),
		Conflict:          domain.ConflictRemoteDeletedLocalModified,
		LocalModifiedUnix: 201,
		BaseModifiedUnix:  110,
	}

	require.Equal(t, expectedConflict, plan.Conflicts[0])
}

func TestPlanner_ContextCanceled_ReturnsError(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	baseSnap := createSnapshot(t, nil)
	localSnap := createSnapshot(t, nil)
	remoteSnap := createSnapshot(t, nil)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, plan)
}

func TestPlanner_FileModifiedOnLocal_RemoteGetsModifyChange(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createRegularFileEntry(t, "/a", "hash-base", 1, 10, 100),
	}

	localFiles := []domain.FileEntry{
		createRegularFileEntry(t, "/a", "hash-local", 1, 20, 140),
	}

	remoteFiles := []domain.FileEntry{
		createRegularFileEntry(t, "/a", "hash-remote", 1, 10, 100),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 0)
	require.Len(t, plan.RemoteChanges, 1)
	require.Len(t, plan.Conflicts, 0)
	require.Equal(t, domain.Modify, plan.RemoteChanges[0].ChangeType)
	require.Equal(t, mustPath(t, "/a"), plan.RemoteChanges[0].OldRelativePath)
}

func TestPlanner_FileModifiedOnRemote_LocalGetsModifyChange(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createRegularFileEntry(t, "/a", "hash-base", 1, 10, 100),
	}

	localFiles := []domain.FileEntry{
		createRegularFileEntry(t, "/a", "hash-local", 1, 10, 100),
	}

	remoteFiles := []domain.FileEntry{
		createRegularFileEntry(t, "/a", "hash-remote", 1, 20, 140),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 1)
	require.Len(t, plan.RemoteChanges, 0)
	require.Len(t, plan.Conflicts, 0)
	require.Equal(t, domain.Modify, plan.LocalChanges[0].ChangeType)
	require.Equal(t, mustPath(t, "/a"), plan.LocalChanges[0].OldRelativePath)
}

func TestPlanner_BothModifiedAtSameTime_ReturnsConflict(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createRegularFileEntry(t, "/a", "hash-base", 1, 10, 100),
	}

	sameModifiedMetadata := createMetadata(20, 140, false)
	localFiles := []domain.FileEntry{
		createFileEntryWithMetadata(t, "/a", "hash-local", 1, sameModifiedMetadata),
	}
	remoteFiles := []domain.FileEntry{
		createFileEntryWithMetadata(t, "/a", "hash-remote", 1, sameModifiedMetadata),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 0)
	require.Len(t, plan.RemoteChanges, 0)
	require.Len(t, plan.Conflicts, 1)
	require.Equal(t, domain.ConflictBothModifiedAtSameTime, plan.Conflicts[0].Conflict)
}

func TestPlanner_FileRenamedOnLocal_RenameChangeOnRemote(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-base", 2, 10, 0),
		createRegularFileEntry(t, "/docs/report.txt", "hash-1", 1, 10, 100),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-local", 2, 20, 0),
		createRegularFileEntry(t, "/docs/final.txt", "hash-1", 1, 20, 100),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-base", 2, 10, 0),
		createRegularFileEntry(t, "/docs/report.txt", "hash-1", 1, 10, 100),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 0)
	require.Len(t, plan.RemoteChanges, 1)
	require.Len(t, plan.Conflicts, 0)

	expectedRemoteChange := domain.RemoteChange{
		OldRelativePath: mustPath(t, "/docs/report.txt"),
		NewRelativePath: mustPath(t, "/docs/final.txt"),
		ChangeType:      domain.Rename,
	}

	require.Equal(t, expectedRemoteChange, plan.RemoteChanges[0])
}

func TestPlanner_FileMovedOnLocal_MoveChangeOnRemote(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-base", 2, 10, 0),
		createRegularFileEntry(t, "/docs/report.txt", "hash-1", 1, 10, 100),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-base", 2, 10, 0),
		createDirectoryEntry(t, "/archive", "dir-archive-local", 2, 20, 0),
		createRegularFileEntry(t, "/archive/report.txt", "hash-1", 1, 20, 100),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-base", 2, 10, 0),
		createRegularFileEntry(t, "/docs/report.txt", "hash-1", 1, 10, 100),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 0)
	require.Len(t, plan.RemoteChanges, 1)
	require.Len(t, plan.Conflicts, 0)

	expectedRemoteChange := domain.RemoteChange{
		OldRelativePath: mustPath(t, "/docs/report.txt"),
		NewRelativePath: mustPath(t, "/archive/report.txt"),
		ChangeType:      domain.Move,
	}

	require.Equal(t, expectedRemoteChange, plan.RemoteChanges[0])
}

func TestPlanner_FileRenamedOnRemote_RenameChangeOnLocal(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-base", 2, 10, 0),
		createRegularFileEntry(t, "/docs/report.txt", "hash-1", 1, 10, 100),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-base", 2, 10, 0),
		createRegularFileEntry(t, "/docs/report.txt", "hash-1", 1, 10, 100),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-remote", 2, 20, 0),
		createRegularFileEntry(t, "/docs/final.txt", "hash-1", 1, 20, 100),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 1)
	require.Len(t, plan.RemoteChanges, 0)
	require.Len(t, plan.Conflicts, 0)

	expectedLocalChange := domain.LocalChange{
		OldRelativePath: mustPath(t, "/docs/report.txt"),
		NewRelativePath: mustPath(t, "/docs/final.txt"),
		ChangeType:      domain.Rename,
	}

	require.Equal(t, expectedLocalChange, plan.LocalChanges[0])
}

func TestPlanner_BothRenamedDifferently_ReturnsConflict(t *testing.T) {
	planner := NewChangesPlannerWithTreeSkip()
	ctx := context.Background()

	baseFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-base", 2, 10, 0),
		createRegularFileEntry(t, "/docs/report.txt", "hash-1", 1, 10, 100),
	}

	localFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-local", 2, 20, 0),
		createRegularFileEntry(t, "/docs/local.txt", "hash-1", 1, 20, 100),
	}

	remoteFiles := []domain.FileEntry{
		createDirectoryEntry(t, "/docs", "dir-docs-remote", 2, 30, 0),
		createRegularFileEntry(t, "/docs/remote.txt", "hash-1", 1, 30, 100),
	}

	baseSnap := createSnapshot(t, baseFiles)
	localSnap := createSnapshot(t, localFiles)
	remoteSnap := createSnapshot(t, remoteFiles)

	plan, err := planner.Plan(ctx, baseSnap, localSnap, remoteSnap)
	require.NoError(t, err)

	require.Len(t, plan.LocalChanges, 0)
	require.Len(t, plan.RemoteChanges, 0)
	require.Len(t, plan.Conflicts, 1)
	require.Equal(t, domain.ConflictLocalRenamedRemoteRenamed, plan.Conflicts[0].Conflict)
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

func createRegularFileEntry(
	t *testing.T,
	relativePath string,
	hash string,
	subtreeSize uint64,
	modifiedUnix uint64,
	sizeBytes uint64,
) domain.FileEntry {
	t.Helper()
	return createFileEntryWithMetadata(t, relativePath, hash, subtreeSize, createMetadata(modifiedUnix, sizeBytes, false))
}

func createDirectoryEntry(
	t *testing.T,
	relativePath string,
	hash string,
	subtreeSize uint64,
	modifiedUnix uint64,
	sizeBytes uint64,
) domain.FileEntry {
	t.Helper()
	return createFileEntryWithMetadata(t, relativePath, hash, subtreeSize, createMetadata(modifiedUnix, sizeBytes, true))
}

func createFileEntryWithMetadata(
	t *testing.T,
	relativePath string,
	hash string,
	subtreeSize uint64,
	metadata domain.FileMetadata,
) domain.FileEntry {
	t.Helper()
	return domain.FileEntry{
		RelativePath: mustPath(t, relativePath),
		SubtreeSize:  subtreeSize,
		FileInfo: domain.FileInfo{
			Hash:     hash,
			Metadata: metadata,
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
