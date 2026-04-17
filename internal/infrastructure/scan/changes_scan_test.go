package scan

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"insync/internal/domain"
)

func TestChangesScannerWithTreeSkip_Scan(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	scanner := NewChangesPlannerWithTreeSkip(ChangesPlannerWithTreeSkipOptions{
		Logger: logger,
	})

	tests := []struct {
		name           string
		baseSnapshot   domain.Snapshot
		localSnapshot  domain.Snapshot
		remoteSnapshot domain.Snapshot
		wantLocal      []domain.LocalChange
		wantRemote     []domain.RemoteChange
		wantConflicts  []domain.Conflict
	}{
		{
			name:           "no changes - identical snapshots",
			baseSnapshot:   newSnapshot(newFile(t, "file1.txt", "hash1", 100, false)),
			localSnapshot:  newSnapshot(newFile(t, "file1.txt", "hash1", 100, false)),
			remoteSnapshot: newSnapshot(newFile(t, "file1.txt", "hash1", 100, false)),
		},
		{
			name:           "new file on local only",
			baseSnapshot:   newSnapshot(),
			localSnapshot:  newSnapshot(newFile(t, "new.txt", "hashA", 101, false)),
			remoteSnapshot: newSnapshot(),
			wantRemote: []domain.RemoteChange{
				newRemoteChange(domain.Create, "", "new.txt", 101),
			},
		},
		{
			name:           "new file on remote only",
			baseSnapshot:   newSnapshot(),
			localSnapshot:  newSnapshot(),
			remoteSnapshot: newSnapshot(newFile(t, "new.txt", "hashB", 101, false)),
			wantLocal: []domain.LocalChange{
				newLocalChange(domain.Create, "", "new.txt", 101),
			},
		},
		{
			name:           "conflict - both created same path different content",
			baseSnapshot:   newSnapshot(),
			localSnapshot:  newSnapshot(newFile(t, "conflict.txt", "hashL", 101, false)),
			remoteSnapshot: newSnapshot(newFile(t, "conflict.txt", "hashR", 101, false)),
			wantConflicts: []domain.Conflict{
				newConflict(
					domain.ConflictBothCreatedAtSamePathConflict,
					"conflict.txt", "conflict.txt", "",
					101, 101, 0,
				),
			},
		},
		{
			name:           "delete on local, unchanged on remote → delete on remote",
			baseSnapshot:   newSnapshot(newFile(t, "del.txt", "hash1", 100, false)),
			localSnapshot:  newSnapshot(),
			remoteSnapshot: newSnapshot(newFile(t, "del.txt", "hash1", 100, false)),
			wantRemote: []domain.RemoteChange{
				newRemoteChange(domain.Delete, "del.txt", "", 100),
			},
		},
		{
			name:           "delete on local + modified on remote → conflict",
			baseSnapshot:   newSnapshot(newFile(t, "mod.txt", "hash1", 100, false)),
			localSnapshot:  newSnapshot(),
			remoteSnapshot: newSnapshot(newFile(t, "mod.txt", "hash2", 102, false)),
			wantConflicts: []domain.Conflict{
				newConflict(
					domain.ConflictLocalDeletedRemoteModified,
					"", "mod.txt", "mod.txt",
					0, 102, 100,
				),
			},
		},
		{
			name:           "modification - local newer",
			baseSnapshot:   newSnapshot(newFile(t, "file.txt", "hash1", 100, false)),
			localSnapshot:  newSnapshot(newFile(t, "file.txt", "hash2", 200, false)),
			remoteSnapshot: newSnapshot(newFile(t, "file.txt", "hash3", 150, false)),
			wantRemote: []domain.RemoteChange{
				newRemoteChange(domain.Modify, "file.txt", "", 200),
			},
		},
		{
			name:           "rename on local only",
			baseSnapshot:   newSnapshot(newFile(t, "old.txt", "hashX", 100, false)),
			localSnapshot:  newSnapshot(newFile(t, "new.txt", "hashX", 100, false)),
			remoteSnapshot: newSnapshot(newFile(t, "old.txt", "hashX", 100, false)),
			wantRemote: []domain.RemoteChange{
				newRemoteChangeRename("old.txt", "new.txt", 100),
			},
		},
		{
			name:           "move on local only",
			baseSnapshot:   newSnapshot(newFile(t, "docs/old.txt", "hashY", 100, false)),
			localSnapshot:  newSnapshot(newFile(t, "archive/old.txt", "hashY", 100, false)),
			remoteSnapshot: newSnapshot(newFile(t, "docs/old.txt", "hashY", 100, false)),
			wantRemote: []domain.RemoteChange{
				newRemoteChangeMove("docs/old.txt", "archive/old.txt", 100),
			},
		},
		{
			name:           "rename + rename conflict",
			baseSnapshot:   newSnapshot(newFile(t, "file.txt", "hashZ", 100, false)),
			localSnapshot:  newSnapshot(newFile(t, "file_local.txt", "hashZ", 100, false)),
			remoteSnapshot: newSnapshot(newFile(t, "file_remote.txt", "hashZ", 100, false)),
			wantConflicts: []domain.Conflict{
				newConflict(
					domain.ConflictLocalRenamedRemoteRenamed,
					"file_local.txt", "file_remote.txt", "file.txt",
					100, 100, 100,
				),
			},
		},
		{
			name: "tree skip - identical directory subtree",
			baseSnapshot: newSnapshot(
				newDir(t, "dir/", "hashD", 3),
				newFile(t, "dir/sub.txt", "hashS", 100, false),
			),
			localSnapshot: newSnapshot(
				newDir(t, "dir/", "hashD", 3),
				newFile(t, "dir/sub.txt", "hashS", 100, false),
			),
			remoteSnapshot: newSnapshot(
				newDir(t, "dir/", "hashD", 3),
				newFile(t, "dir/sub.txt", "hashS", 100, false),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := scanner.Plan(t.Context(), tt.baseSnapshot, tt.localSnapshot, tt.remoteSnapshot)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.wantLocal, plan.LocalChanges, "local changes mismatch")
			assert.Equal(t, tt.wantRemote, plan.RemoteChanges, "remote changes mismatch")
			assert.Equal(t, tt.wantConflicts, plan.Conflicts, "conflicts mismatch")
		})
	}
}

func newSnapshot(files ...domain.FileEntry) domain.Snapshot {
	return domain.NewSnapshot(files)
}

func newFile(t *testing.T, relPath, hash string, modified uint64, isDir bool) domain.FileEntry {
	path, err := domain.NewPath(relPath)
	require.NoError(t, err)

	fi := domain.FileInfo{
		Hash: hash,
		Metadata: domain.FileMetadata{
			ModifiedUnix: modified,
			IsDirectory:  isDir,
		},
	}

	entry, err := domain.NewFileEntry(path, 1, fi)
	require.NoError(t, err)

	return entry
}

func newDir(t *testing.T, relPath, hash string, subtreeSize uint64) domain.FileEntry {
	path, err := domain.NewPath(relPath)
	require.NoError(t, err)

	fi := domain.FileInfo{
		Hash: hash,
		Metadata: domain.FileMetadata{
			ModifiedUnix: 100,
			IsDirectory:  true,
		},
	}

	entry, err := domain.NewFileEntry(path, subtreeSize, fi)
	require.NoError(t, err)

	return entry
}

func newLocalChange(changeType domain.SyncChangeType, oldPath, newPath string, modifiedUnix uint64) domain.LocalChange {
	return domain.SyncChange{
		OldRelativePath: mustPath(oldPath),
		NewRelativePath: mustPath(newPath),
		ChangeType:      changeType,
		ModifiedUnix:    modifiedUnix,
	}.ToLocalChange()
}

func newRemoteChange(changeType domain.SyncChangeType, oldPath, newPath string, modifiedUnix uint64) domain.RemoteChange {
	return domain.SyncChange{
		OldRelativePath: mustPath(oldPath),
		NewRelativePath: mustPath(newPath),
		ChangeType:      changeType,
		ModifiedUnix:    modifiedUnix,
	}.ToRemoteChange()
}

func newRemoteChangeRename(oldPath, newPath string, modifiedUnix uint64) domain.RemoteChange {
	return newRemoteChange(domain.Rename, oldPath, newPath, modifiedUnix)
}

func newRemoteChangeMove(oldPath, newPath string, modifiedUnix uint64) domain.RemoteChange {
	return newRemoteChange(domain.Move, oldPath, newPath, modifiedUnix)
}

func newConflict(
	conflictType domain.ConflictType,
	localPath, remotePath, basePath string,
	localModifiedUnix, remoteModifiedUnix, baseModifiedUnix uint64,
) domain.Conflict {
	return domain.Conflict{
		LocalRelativePath:  mustPath(localPath),
		RemoteRelativePath: mustPath(remotePath),
		BaseRelativePath:   mustPath(basePath),
		LocalModifiedUnix:  localModifiedUnix,
		RemoteModifiedUnix: remoteModifiedUnix,
		BaseModifiedUnix:   baseModifiedUnix,
		Conflict:           conflictType,
	}
}

func mustPath(pathString string) domain.Path {
	if pathString == "" {
		return ""
	}
	parsedPath, err := domain.NewPath(pathString)
	if err != nil {
		panic(err)
	}
	return parsedPath
}
