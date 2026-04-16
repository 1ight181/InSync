package scan

import (
	"context"
	"insync/internal/domain"
	"log/slog"
	"path/filepath"
)

const sentinel = "\xFF\xFF"

type ChangesScannerWithTreeSkip struct {
	logger    *slog.Logger
	loggerCtx context.Context
}

type ChangesScannerWithTreeSkipOptions struct {
	Logger *slog.Logger
}

func NewChangesScannerWithTreeSkip(opts ChangesScannerWithTreeSkipOptions) *ChangesScannerWithTreeSkip {
	if opts.Logger == nil {
		panic("Logger must be provided to ChangesScannerWithTreeSkip")
	}
	return &ChangesScannerWithTreeSkip{
		logger:    opts.Logger,
		loggerCtx: context.Background(),
	}
}

type potentialChange struct {
	Changetype domain.SyncChangeType
	Path       domain.Path
	Hash       string
	Modify     uint64
}

func (s *ChangesScannerWithTreeSkip) Scan(
	baseSnapshot, localSnapshot, remoteSnapshot domain.Snapshot,
) ([]domain.LocalChange, []domain.RemoteChange, []domain.Conflict) {

	var localChanges []domain.LocalChange
	var remoteChanges []domain.RemoteChange
	var conflicts []domain.Conflict

	var potentialLocalChanges []potentialChange
	var potentialRemoteChanges []potentialChange

	var baseIdx, localIdx, remoteIdx int

	base := baseSnapshot.Files
	local := localSnapshot.Files
	remote := remoteSnapshot.Files

	for baseIdx < len(base) || localIdx < len(local) || remoteIdx < len(remote) {
		pBase := s.pathAt(base, baseIdx)
		pLocal := s.pathAt(local, localIdx)
		pRemote := s.pathAt(remote, remoteIdx)

		minPath := s.minString(pBase, pLocal, pRemote)

		var baseEntry, localEntry, remoteEntry *domain.FileEntry

		if baseIdx < len(base) && base[baseIdx].RelativePath.String() == minPath {
			baseEntry = &base[baseIdx]
		}
		if localIdx < len(local) && local[localIdx].RelativePath.String() == minPath {
			localEntry = &local[localIdx]
		}
		if remoteIdx < len(remote) && remote[remoteIdx].RelativePath.String() == minPath {
			remoteEntry = &remote[remoteIdx]
		}

		if s.canSkipSubtreeAt(baseEntry, localEntry, remoteEntry, minPath) {
			s.skipSubtree(int(baseSnapshot.Files[baseIdx].SubtreeSize), &baseIdx)
			s.skipSubtree(int(localSnapshot.Files[localIdx].SubtreeSize), &localIdx)
			s.skipSubtree(int(remoteSnapshot.Files[remoteIdx].SubtreeSize), &remoteIdx)
			continue
		}

		localChange, remoteChange, conflict, potentialLocalChange, potentialRemoteChange := s.processEntry(
			baseEntry, localEntry, remoteEntry,
		)

		switch {
		case localChange != nil:
			localChanges = append(localChanges, *localChange)
		case remoteChange != nil:
			remoteChanges = append(remoteChanges, *remoteChange)
		case conflict != nil:
			conflicts = append(conflicts, *conflict)
		case potentialLocalChange != nil:
			potentialLocalChanges = append(potentialLocalChanges, *potentialLocalChange)
		}
		if potentialRemoteChange != nil {
			potentialRemoteChanges = append(potentialRemoteChanges, *potentialRemoteChange)
		}

		if baseEntry != nil && baseEntry.RelativePath.String() == minPath {
			baseIdx++
		}
		if localEntry != nil && localEntry.RelativePath.String() == minPath {
			localIdx++
		}
		if remoteEntry != nil && remoteEntry.RelativePath.String() == minPath {
			remoteIdx++
		}
	}

	appliedLocalChanges, appliedRemoteChanges, appliedConflicts := s.detectRenamesAndMoves(potentialLocalChanges, potentialRemoteChanges, base)
	localChanges = append(localChanges, appliedLocalChanges...)
	remoteChanges = append(remoteChanges, appliedRemoteChanges...)
	conflicts = append(conflicts, appliedConflicts...)

	return localChanges, remoteChanges, conflicts
}

func (s *ChangesScannerWithTreeSkip) processEntry(
	baseEntry, localEntry, remoteEntry *domain.FileEntry,
) (
	*domain.LocalChange,
	*domain.RemoteChange,
	*domain.Conflict,
	*potentialChange,
	*potentialChange,
) {
	switch {
	case s.isModified(baseEntry, localEntry, remoteEntry):
		localChange, remoteChange, conflict := s.handleModification(localEntry, remoteEntry)
		return localChange, remoteChange, conflict, nil, nil

	case s.isNewFile(baseEntry, localEntry, remoteEntry):
		potentialLocalChange, potentialRemoteChange, conflict := s.handleCreation(localEntry, remoteEntry)
		return nil, nil, conflict, potentialLocalChange, potentialRemoteChange

	case s.isDeleted(baseEntry, localEntry, remoteEntry):
		conflict, potentialLocalChange, potentialRemoteChange := s.handleDeletion(baseEntry, localEntry, remoteEntry)
		return nil, nil, conflict, potentialLocalChange, potentialRemoteChange
	}

	return nil, nil, nil, nil, nil
}

func (s *ChangesScannerWithTreeSkip) handleDeletion(
	baseEntry, localEntry, remoteEntry *domain.FileEntry,
) (
	*domain.Conflict,
	*potentialChange,
	*potentialChange,
) {

	if localEntry == nil && remoteEntry != nil {
		// Конфликт удаления - удалено на local, но изменено на remote
		if remoteEntry.FileInfo.Hash != baseEntry.FileInfo.Hash {
			return &domain.Conflict{
				RemoteRelativePath: remoteEntry.RelativePath,
				BaseRelativePath:   baseEntry.RelativePath,
				RemoteModifiedUnix: remoteEntry.FileInfo.Metadata.ModifiedUnix,
				BaseModifiedUnix:   baseEntry.FileInfo.Metadata.ModifiedUnix,
				Conflict:           domain.ConflictLocalDeletedRemoteModified,
			}, nil, nil
		}

		// Удалено локально - нужно удалить на remote
		return nil, nil, &potentialChange{
			Changetype: domain.Delete,
			Hash:       baseEntry.FileInfo.Hash,
			Modify:     baseEntry.FileInfo.Metadata.ModifiedUnix,
			Path:       baseEntry.RelativePath,
		}
	}

	if remoteEntry == nil && localEntry != nil {
		// Конфликт удаления - удалено на remote, но изменено на local
		if localEntry.FileInfo.Hash != baseEntry.FileInfo.Hash {
			return &domain.Conflict{
				LocalRelativePath: localEntry.RelativePath,
				BaseRelativePath:  baseEntry.RelativePath,
				LocalModifiedUnix: localEntry.FileInfo.Metadata.ModifiedUnix,
				BaseModifiedUnix:  baseEntry.FileInfo.Metadata.ModifiedUnix,
				Conflict:          domain.ConflictRemoteDeletedLocalModified,
			}, nil, nil
		}

		// Удалено на remote - нужно удалить на local
		return nil, &potentialChange{
			Changetype: domain.Delete,
			Hash:       baseEntry.FileInfo.Hash,
			Modify:     baseEntry.FileInfo.Metadata.ModifiedUnix,
			Path:       baseEntry.RelativePath,
		}, nil
	}

	return nil, nil, nil
}

func (s *ChangesScannerWithTreeSkip) handleCreation(
	localEntry, remoteEntry *domain.FileEntry,
) (*potentialChange, *potentialChange, *domain.Conflict) {
	// создать на remote
	if localEntry != nil && remoteEntry == nil {
		return nil, &potentialChange{
			Changetype: domain.Create,
			Hash:       localEntry.FileInfo.Hash,
			Modify:     localEntry.FileInfo.Metadata.ModifiedUnix,
			Path:       localEntry.RelativePath,
		}, nil
	}

	// создать на local
	if remoteEntry != nil && localEntry == nil {
		return &potentialChange{
			Changetype: domain.Create,
			Hash:       remoteEntry.FileInfo.Hash,
			Modify:     remoteEntry.FileInfo.Metadata.ModifiedUnix,
			Path:       remoteEntry.RelativePath,
		}, nil, nil
	}

	if localEntry != nil && remoteEntry != nil {
		// Конфликт создания - оба создали файл с разным содержимым
		if localEntry.FileInfo.Hash != remoteEntry.FileInfo.Hash {
			return nil, nil, &domain.Conflict{
				LocalRelativePath:  localEntry.RelativePath,
				RemoteRelativePath: remoteEntry.RelativePath,
				LocalModifiedUnix:  localEntry.FileInfo.Metadata.ModifiedUnix,
				RemoteModifiedUnix: remoteEntry.FileInfo.Metadata.ModifiedUnix,
				Conflict:           domain.ConflictBothCreatedAtSamePathConflict,
			}
		}
	}
	return nil, nil, nil
}

func (s *ChangesScannerWithTreeSkip) handleModification(
	local, remote *domain.FileEntry,
) (*domain.LocalChange, *domain.RemoteChange, *domain.Conflict) {
	if local.FileInfo.Metadata.ModifiedUnix > remote.FileInfo.Metadata.ModifiedUnix {
		return nil, &domain.RemoteChange{
			OldRelativePath: local.RelativePath,
			ChangeType:      domain.Modify,
			ModifiedUnix:    local.FileInfo.Metadata.ModifiedUnix,
		}, nil
	} else if local.FileInfo.Metadata.ModifiedUnix == remote.FileInfo.Metadata.ModifiedUnix {
		return nil, nil, &domain.Conflict{
			LocalRelativePath:  local.RelativePath,
			RemoteRelativePath: remote.RelativePath,
			LocalModifiedUnix:  local.FileInfo.Metadata.ModifiedUnix,
			RemoteModifiedUnix: remote.FileInfo.Metadata.ModifiedUnix,
			Conflict:           domain.ConflictBothModifiedConflictAtSameTime,
		}
	}
	return &domain.LocalChange{
		OldRelativePath: remote.RelativePath,
		ChangeType:      domain.Modify,
		ModifiedUnix:    remote.FileInfo.Metadata.ModifiedUnix,
	}, nil, nil
}
func (s *ChangesScannerWithTreeSkip) detectRenamesAndMoves(
	potentialLocalChanges, potentialRemoteChanges []potentialChange, baseFiles []domain.FileEntry,
) ([]domain.LocalChange, []domain.RemoteChange, []domain.Conflict) {
	localChanges := make([]domain.LocalChange, 0, len(potentialLocalChanges))
	remoteChanges := make([]domain.RemoteChange, 0, len(potentialRemoteChanges))
	conflicts := make([]domain.Conflict, 0)

	// hash -> изменения, которые нужно применить на local
	localByHash := make(map[string][]potentialChange)
	// hash -> изменения, которые нужно применить на remote
	remoteByHash := make(map[string][]potentialChange)

	for _, change := range potentialLocalChanges {
		localByHash[change.Hash] = append(localByHash[change.Hash], change)
	}
	for _, change := range potentialRemoteChanges {
		remoteByHash[change.Hash] = append(remoteByHash[change.Hash], change)
	}

	baseByHash := make(map[string]domain.FileEntry, len(baseFiles))
	for _, entry := range baseFiles {
		baseByHash[entry.FileInfo.Hash] = entry
	}

	extractPair := func(changes []potentialChange) (deleteChange, createChange *potentialChange) {
		for i := range changes {
			switch changes[i].Changetype {
			case domain.Delete:
				if deleteChange == nil {
					deleteChange = &changes[i]
				}
			case domain.Create:
				if createChange == nil {
					createChange = &changes[i]
				}
			}
		}
		return deleteChange, createChange
	}

	buildLocalChange := func(baseEntry domain.FileEntry, newPath domain.Path, modifiedUnix uint64) domain.LocalChange {
		baseDir, baseName := filepath.Split(baseEntry.RelativePath.String())
		newDir, newName := filepath.Split(newPath.String())

		if baseDir == newDir && baseName != newName {
			return domain.LocalChange{
				OldRelativePath: baseEntry.RelativePath,
				NewRelativePath: newPath,
				ChangeType:      domain.Rename,
				ModifiedUnix:    modifiedUnix,
			}
		}

		return domain.LocalChange{
			OldRelativePath: baseEntry.RelativePath,
			NewRelativePath: newPath,
			ChangeType:      domain.Move,
			ModifiedUnix:    modifiedUnix,
		}
	}

	buildRemoteChange := func(baseEntry domain.FileEntry, newPath domain.Path, modifiedUnix uint64) domain.RemoteChange {
		baseDir, baseName := filepath.Split(baseEntry.RelativePath.String())
		newDir, newName := filepath.Split(newPath.String())

		if baseDir == newDir && baseName != newName {
			return domain.RemoteChange{
				OldRelativePath: baseEntry.RelativePath,
				NewRelativePath: newPath,
				ChangeType:      domain.Rename,
				ModifiedUnix:    modifiedUnix,
			}
		}

		return domain.RemoteChange{
			OldRelativePath: baseEntry.RelativePath,
			NewRelativePath: newPath,
			ChangeType:      domain.Move,
			ModifiedUnix:    modifiedUnix,
		}
	}

	buildConflict := func(
		baseEntry domain.FileEntry,
		localCreate, remoteCreate *potentialChange,
	) domain.Conflict {
		baseDir, baseName := filepath.Split(baseEntry.RelativePath.String())
		localDir, localName := filepath.Split(localCreate.Path.String())
		remoteDir, remoteName := filepath.Split(remoteCreate.Path.String())

		conflictType := domain.ConflictLocalMovedRemoteMoved
		if baseDir == localDir && baseName != localName &&
			baseDir == remoteDir && baseName != remoteName {
			conflictType = domain.ConflictLocalRenamedRemoteRenamed
		}

		return domain.Conflict{
			LocalRelativePath:  localCreate.Path,
			RemoteRelativePath: remoteCreate.Path,
			BaseRelativePath:   baseEntry.RelativePath,
			LocalModifiedUnix:  localCreate.Modify,
			RemoteModifiedUnix: remoteCreate.Modify,
			BaseModifiedUnix:   baseEntry.FileInfo.Metadata.ModifiedUnix,
			Conflict:           conflictType,
		}
	}

	processedHashes := make(map[string]struct{}, len(localByHash)+len(remoteByHash))
	for hash := range localByHash {
		processedHashes[hash] = struct{}{}
	}
	for hash := range remoteByHash {
		processedHashes[hash] = struct{}{}
	}

	for hash := range processedHashes {
		baseEntry, hasBase := baseByHash[hash]
		if !hasBase {
			continue
		}

		_, localCreate := extractPair(localByHash[hash])
		_, remoteCreate := extractPair(remoteByHash[hash])

		switch {
		case localCreate != nil && remoteCreate != nil:
			// обе стороны создали новый путь для одного hash → rename vs rename
			if localCreate.Path.String() == remoteCreate.Path.String() {
				// одинаковый результат → no-op
				delete(localByHash, hash)
				delete(remoteByHash, hash)
				continue
			}

			conflicts = append(conflicts, buildConflict(baseEntry, remoteCreate, localCreate))
			delete(localByHash, hash)
			delete(remoteByHash, hash)
			continue

		case localCreate != nil:
			// изменение произошло на remote → применяем на local
			localChanges = append(localChanges,
				buildLocalChange(baseEntry, localCreate.Path, localCreate.Modify),
			)
			delete(localByHash, hash)
			continue

		case remoteCreate != nil:
			// изменение произошло на local → применяем на remote
			remoteChanges = append(remoteChanges,
				buildRemoteChange(baseEntry, remoteCreate.Path, remoteCreate.Modify),
			)
			delete(remoteByHash, hash)
			continue
		}
	}

	// Остатки без пары — обычные create/delete.
	for _, changes := range remoteByHash {
		for _, change := range changes {
			switch change.Changetype {
			case domain.Create:
				remoteChanges = append(remoteChanges, domain.RemoteChange{
					ChangeType:      domain.Create,
					NewRelativePath: change.Path,
					ModifiedUnix:    change.Modify,
				})
			case domain.Delete:
				remoteChanges = append(remoteChanges, domain.RemoteChange{
					ChangeType:      domain.Delete,
					OldRelativePath: change.Path,
					ModifiedUnix:    change.Modify,
				})
			}
		}
	}

	for _, changes := range localByHash {
		for _, change := range changes {
			switch change.Changetype {
			case domain.Create:
				localChanges = append(localChanges, domain.LocalChange{
					ChangeType:      domain.Create,
					NewRelativePath: change.Path,
					ModifiedUnix:    change.Modify,
				})
			case domain.Delete:
				localChanges = append(localChanges, domain.LocalChange{
					ChangeType:      domain.Delete,
					OldRelativePath: change.Path,
					ModifiedUnix:    change.Modify,
				})
			}
		}
	}

	return localChanges, remoteChanges, conflicts
}

func (s *ChangesScannerWithTreeSkip) isRenamedRenamedConflict(
	baseEntryDir, localDir, baseEntryFileName, localFileName, remoteDir, remoteFileName string,
) bool {
	return baseEntryDir == localDir && baseEntryFileName != localFileName &&
		baseEntryDir == remoteDir && baseEntryFileName != remoteFileName
}

func (s *ChangesScannerWithTreeSkip) isMovedMovedConflict(
	baseEntryDir, localDir, remoteDir string,
) bool {
	return baseEntryDir != localDir &&
		baseEntryDir != remoteDir &&
		localDir != remoteDir
}

func (s *ChangesScannerWithTreeSkip) pathAt(entries []domain.FileEntry, idx int) string {
	if idx >= len(entries) {
		return sentinel
	}
	return entries[idx].RelativePath.String()
}

func (s *ChangesScannerWithTreeSkip) minString(a, b, c string) string {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

func (s *ChangesScannerWithTreeSkip) canSkipSubtreeAt(b, l, r *domain.FileEntry, currentPath string) bool {
	if b == nil || l == nil || r == nil {
		return false
	}
	if b.RelativePath.String() != currentPath ||
		l.RelativePath.String() != currentPath ||
		r.RelativePath.String() != currentPath {
		return false
	}
	return b.FileInfo.Hash == l.FileInfo.Hash && b.FileInfo.Hash == r.FileInfo.Hash
}

func (s *ChangesScannerWithTreeSkip) skipSubtree(subtreeSize int, idx *int) {
	*idx += subtreeSize
}

func (s *ChangesScannerWithTreeSkip) isDeleted(base, local, remote *domain.FileEntry) bool {
	return base != nil && (local == nil || remote == nil)
}

func (s *ChangesScannerWithTreeSkip) isNewFile(base, local, remote *domain.FileEntry) bool {
	if base != nil {
		return false
	}
	return local != nil || remote != nil
}

func (s *ChangesScannerWithTreeSkip) isModified(base, local, remote *domain.FileEntry) bool {
	return base != nil && local != nil && remote != nil && local.FileInfo.Hash != remote.FileInfo.Hash
}
