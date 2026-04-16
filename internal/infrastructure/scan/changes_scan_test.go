package scan

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"insync/internal/domain"
)

func TestChangesScannerWithTreeSkip_Scan(t *testing.T) {
	logger := slog.New(slog.DiscardHandler) // тихий логгер для тестов

	scanner := NewChangesScannerWithTreeSkip(ChangesScannerWithTreeSkipOptions{
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
			baseSnapshot:   newSnapshot(100, newFile("file1.txt", "hash1", 100, false)),
			localSnapshot:  newSnapshot(100, newFile("file1.txt", "hash1", 100, false)),
			remoteSnapshot: newSnapshot(100, newFile("file1.txt", "hash1", 100, false)),
			wantLocal:      nil,
			wantRemote:     nil,
			wantConflicts:  nil,
		},
		{
			name:           "new file on local only",
			baseSnapshot:   newSnapshot(100),
			localSnapshot:  newSnapshot(101, newFile("new.txt", "hashA", 101, false)),
			remoteSnapshot: newSnapshot(100),
			wantLocal:      nil,
			wantRemote: []domain.RemoteChange{
				newRemoteChange(domain.Create, "", "new.txt", 101),
			},
			wantConflicts: nil,
		},
		{
			name:           "new file on remote only",
			baseSnapshot:   newSnapshot(100),
			localSnapshot:  newSnapshot(100),
			remoteSnapshot: newSnapshot(101, newFile("new.txt", "hashB", 101, false)),
			wantLocal: []domain.LocalChange{
				newLocalChange(domain.Create, "", "new.txt", 101),
			},
			wantRemote:    nil,
			wantConflicts: nil,
		},
		{
			name:           "conflict - both created same path different content",
			baseSnapshot:   newSnapshot(100),
			localSnapshot:  newSnapshot(101, newFile("conflict.txt", "hashL", 101, false)),
			remoteSnapshot: newSnapshot(101, newFile("conflict.txt", "hashR", 101, false)),
			wantLocal:      nil,
			wantRemote:     nil,
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
			baseSnapshot:   newSnapshot(100, newFile("del.txt", "hash1", 100, false)),
			localSnapshot:  newSnapshot(101),
			remoteSnapshot: newSnapshot(100, newFile("del.txt", "hash1", 100, false)),
			wantLocal:      nil,
			wantRemote: []domain.RemoteChange{
				newRemoteChange(domain.Delete, "del.txt", "", 100),
			},
			wantConflicts: nil,
		},
		{
			name:           "delete on local + modified on remote → conflict",
			baseSnapshot:   newSnapshot(100, newFile("mod.txt", "hash1", 100, false)),
			localSnapshot:  newSnapshot(101),
			remoteSnapshot: newSnapshot(102, newFile("mod.txt", "hash2", 102, false)),
			wantLocal:      nil,
			wantRemote:     nil,
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
			baseSnapshot:   newSnapshot(100, newFile("file.txt", "hash1", 100, false)),
			localSnapshot:  newSnapshot(200, newFile("file.txt", "hash2", 200, false)),
			remoteSnapshot: newSnapshot(150, newFile("file.txt", "hash3", 150, false)),
			wantLocal:      nil,
			wantRemote: []domain.RemoteChange{
				newRemoteChange(domain.Modify, "file.txt", "", 200),
			},
			wantConflicts: nil,
		},
		{
			name:           "rename on local only",
			baseSnapshot:   newSnapshot(100, newFile("old.txt", "hashX", 100, false)),
			localSnapshot:  newSnapshot(100, newFile("new.txt", "hashX", 100, false)),
			remoteSnapshot: newSnapshot(100, newFile("old.txt", "hashX", 100, false)),
			wantLocal:      nil,
			wantRemote: []domain.RemoteChange{
				newRemoteChangeRename("old.txt", "new.txt", 100),
			},
			wantConflicts: nil,
		},
		{
			name:           "move on local only",
			baseSnapshot:   newSnapshot(100, newFile("docs/old.txt", "hashY", 100, false)),
			localSnapshot:  newSnapshot(100, newFile("archive/old.txt", "hashY", 100, false)),
			remoteSnapshot: newSnapshot(100, newFile("docs/old.txt", "hashY", 100, false)),
			wantLocal:      nil,
			wantRemote: []domain.RemoteChange{
				newRemoteChangeMove("docs/old.txt", "archive/old.txt", 100),
			},
			wantConflicts: nil,
		},
		{
			name:           "rename + rename conflict",
			baseSnapshot:   newSnapshot(100, newFile("file.txt", "hashZ", 100, false)),
			localSnapshot:  newSnapshot(101, newFile("file_local.txt", "hashZ", 100, false)),
			remoteSnapshot: newSnapshot(101, newFile("file_remote.txt", "hashZ", 100, false)),
			wantLocal:      nil,
			wantRemote:     nil,
			wantConflicts: []domain.Conflict{
				newConflict(
					domain.ConflictLocalRenamedRemoteRenamed,
					"file_local.txt", "file_remote.txt", "file.txt",
					100, 100, 100,
				),
			},
		},
		{
			name:           "tree skip - identical directory subtree",
			baseSnapshot:   newSnapshot(100, newDir("dir/", "hashD", 3), newFile("dir/sub.txt", "hashS", 100, false)),
			localSnapshot:  newSnapshot(100, newDir("dir/", "hashD", 3), newFile("dir/sub.txt", "hashS", 100, false)),
			remoteSnapshot: newSnapshot(100, newDir("dir/", "hashD", 3), newFile("dir/sub.txt", "hashS", 100, false)),
			wantLocal:      nil,
			wantRemote:     nil,
			wantConflicts:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localCh, remoteCh, conflicts := scanner.Scan(tt.baseSnapshot, tt.localSnapshot, tt.remoteSnapshot)

			assert.Equal(t, tt.wantLocal, localCh, "local changes mismatch")
			assert.Equal(t, tt.wantRemote, remoteCh, "remote changes mismatch")
			assert.Equal(t, tt.wantConflicts, conflicts, "conflicts mismatch")
		})
	}
}

// ==================== Вспомогательные конструкторы ====================

func newSnapshot(unixTime uint64, files ...domain.FileEntry) domain.Snapshot {
	return domain.NewSnapshot(unixTime, files)
}

func newFile(relPath, hash string, modified uint64, isDir bool) domain.FileEntry {
	path, err := domain.NewPath(relPath)
	require.NoError(nil, err) // в тесте паника при ошибке — нормально

	fi := domain.FileInfo{
		Hash:     hash,
		Metadata: domain.FileMetadata{ModifiedUnix: modified, IsDirectory: isDir},
		// RootName и SizeBytes можно оставить пустыми/0, если не используются в сравнении
	}

	entry, err := domain.NewFileEntry(path, 1, fi) // subtreeSize = 1 для файла
	require.NoError(nil, err)
	return entry
}

func newDir(relPath, hash string, subtreeSize uint64) domain.FileEntry {
	path, _ := domain.NewPath(relPath)
	fi := domain.FileInfo{
		Hash:     hash,
		Metadata: domain.FileMetadata{ModifiedUnix: 100, IsDirectory: true},
	}
	entry, _ := domain.NewFileEntry(path, subtreeSize, fi)
	return entry
}

func newLocalChange(ct domain.SyncChangeType, old, new string, mod uint64) domain.LocalChange {
	return domain.SyncChange{
		OldRelativePath: mustPath(old),
		NewRelativePath: mustPath(new),
		ChangeType:      ct,
		ModifiedUnix:    mod,
	}.ToLocalChange()
}

func newRemoteChange(ct domain.SyncChangeType, old, new string, mod uint64) domain.RemoteChange {
	return domain.SyncChange{
		OldRelativePath: mustPath(old),
		NewRelativePath: mustPath(new),
		ChangeType:      ct,
		ModifiedUnix:    mod,
	}.ToRemoteChange()
}

func newRemoteChangeRename(old, new string, mod uint64) domain.RemoteChange {
	return newRemoteChange(domain.Rename, old, new, mod)
}

func newRemoteChangeMove(old, new string, mod uint64) domain.RemoteChange {
	return newRemoteChange(domain.Move, old, new, mod)
}

func newConflict(
	typ domain.ConflictType,
	localP, remoteP, baseP string,
	localMod, remoteMod, baseMod uint64,
) domain.Conflict {
	return domain.Conflict{
		LocalRelativePath:  mustPath(localP),
		RemoteRelativePath: mustPath(remoteP),
		BaseRelativePath:   mustPath(baseP),
		LocalModifiedUnix:  localMod,
		RemoteModifiedUnix: remoteMod,
		BaseModifiedUnix:   baseMod,
		Conflict:           typ,
	}
}

func mustPath(s string) domain.Path {
	if s == "" {
		return ""
	}
	p, err := domain.NewPath(s)
	if err != nil {
		panic(err)
	}
	return p
}
