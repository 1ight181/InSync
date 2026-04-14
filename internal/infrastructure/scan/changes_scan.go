package scan

import (
	"context"
	"insync/internal/domain"
	"log/slog"
	"path/filepath"
	"strings"
)

type ChangesScanner struct {
	logger    *slog.Logger
	loggerCtx context.Context
}

type ChangesScannerOptions struct {
	Logger *slog.Logger
}

type potential struct {
	Hash     string
	Path     domain.Path
	Modified uint64
}

func NewChangesScanner(opts ChangesScannerOptions) *ChangesScanner {
	if opts.Logger == nil {
		panic("Все поля ChangesScanner должны быть заполнены")
	}
	loggerCtx := context.Background()
	return &ChangesScanner{
		logger:    opts.Logger,
		loggerCtx: loggerCtx,
	}
}
func (s *ChangesScanner) Scan(localSnapshot, remoteSnapshot domain.Snapshot) []domain.SyncChange {
	changes := make([]domain.SyncChange, 0, 64)

	var potentialCreates []potential
	var potentialDeletes []potential

	i, j := 0, 0
	for i < len(localSnapshot.Files) && j < len(remoteSnapshot.Files) {
		localEntry := &localSnapshot.Files[i]
		remoteEntry := &remoteSnapshot.Files[j]

		cmp := strings.Compare(
			localEntry.RelativePath.String(),
			remoteEntry.RelativePath.String(),
		)

		switch {
		case cmp == 0: // одинаковый путь
			if s.canSkipSubtree(localEntry, remoteEntry) {
				i += int(localEntry.SubtreeSize)
				j += int(remoteEntry.SubtreeSize)
				continue
			}

			if localEntry.FileInfo.Hash != remoteEntry.FileInfo.Hash {
				changes = append(changes, domain.SyncChange{
					ChangeType:      domain.Modified,
					RootName:        localSnapshot.RootName,
					OldRelativePath: localEntry.RelativePath,
					ModifiedUnix:    localEntry.FileInfo.Metadata.ModifiedUnix,
				})
			}

			i++
			j++

		case cmp < 0: // есть в local, нет в remote - потенциальный Created
			potentialCreates = append(potentialCreates, potential{
				Hash:     localEntry.FileInfo.Hash,
				Path:     localEntry.RelativePath,
				Modified: localEntry.FileInfo.Metadata.ModifiedUnix,
			})
			i += int(localEntry.SubtreeSize)

		default: // cmp > 0: есть в remote, нет в local - потенциальный Deleted
			potentialDeletes = append(potentialDeletes, potential{
				Hash:     remoteEntry.FileInfo.Hash,
				Path:     remoteEntry.RelativePath,
				Modified: remoteEntry.FileInfo.Metadata.ModifiedUnix,
			})
			j += int(remoteEntry.SubtreeSize)
		}
	}

	// остаток local - Created
	for i < len(localSnapshot.Files) {
		entry := &localSnapshot.Files[i]
		potentialCreates = append(potentialCreates, potential{
			Hash:     entry.FileInfo.Hash,
			Path:     entry.RelativePath,
			Modified: entry.FileInfo.Metadata.ModifiedUnix,
		})
		i += int(entry.SubtreeSize)
	}

	// остаток remote - Deleted
	for j < len(remoteSnapshot.Files) {
		entry := &remoteSnapshot.Files[j]
		potentialDeletes = append(potentialDeletes, potential{
			Hash:     entry.FileInfo.Hash,
			Path:     entry.RelativePath,
			Modified: entry.FileInfo.Metadata.ModifiedUnix,
		})
		j += int(entry.SubtreeSize)
	}

	// map[hash] -> create
	createMap := make(map[string]potential, len(potentialCreates))
	for _, c := range potentialCreates {
		createMap[c.Hash] = c
	}

	for _, del := range potentialDeletes {
		if create, ok := createMap[del.Hash]; ok {
			oldParent := filepath.Dir(del.Path.String())
			newParent := filepath.Dir(create.Path.String())

			if oldParent == newParent {
				// Renamed (имя изменилось, папка та же)
				changes = append(changes, domain.SyncChange{
					ChangeType:      domain.Renamed,
					RootName:        localSnapshot.RootName,
					OldRelativePath: del.Path,
					NewRelativePath: create.Path,
					ModifiedUnix:    create.Modified,
				})
			} else {
				// Moved (перемещён в другую папку)
				changes = append(changes, domain.SyncChange{
					ChangeType:      domain.Moved,
					RootName:        localSnapshot.RootName,
					OldRelativePath: del.Path,
					NewRelativePath: create.Path,
					ModifiedUnix:    create.Modified,
				})
			}
			delete(createMap, del.Hash)
			continue
		}

		// не нашли пару - Deleted
		changes = append(changes, domain.SyncChange{
			ChangeType:      domain.Deleted,
			RootName:        remoteSnapshot.RootName,
			OldRelativePath: del.Path,
			ModifiedUnix:    del.Modified,
		})
	}

	// оставшиеся creates без пары - Created
	for _, c := range createMap {
		changes = append(changes, domain.SyncChange{
			ChangeType:      domain.Created,
			RootName:        localSnapshot.RootName,
			NewRelativePath: c.Path,
			ModifiedUnix:    c.Modified,
		})
	}

	return changes
}

func (s *ChangesScanner) canSkipSubtree(localEntry, remoteEntry *domain.FileEntry) bool {
	if localEntry.SubtreeSize != remoteEntry.SubtreeSize {
		return false
	}

	return localEntry.FileInfo.Metadata.ModifiedUnix == remoteEntry.FileInfo.Metadata.ModifiedUnix &&
		localEntry.FileInfo.Metadata.SizeBytes == remoteEntry.FileInfo.Metadata.SizeBytes &&
		localEntry.FileInfo.Hash == remoteEntry.FileInfo.Hash
}
