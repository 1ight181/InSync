package scan

import (
	"context"
	"fmt"
	"insync/internal/domain"
	"path/filepath"
)

const sentinel = "\xFF\xFF"

type ChangesPlannerWithTreeSkip struct{}

func NewChangesPlannerWithTreeSkip() *ChangesPlannerWithTreeSkip {
	return &ChangesPlannerWithTreeSkip{}
}

// potentialChange отражает изменение которое потенциально требуется совершить
//
// Используется для обнаружения rename/move и связанных конфликтов
type potentialChange struct {
	Changetype domain.SyncChangeType
	Path       domain.Path
	Hash       string
	Modify     uint64
}

// Единственная возвращаемая ошибка - ошибка по контексту.
func (s *ChangesPlannerWithTreeSkip) Plan(
	ctx context.Context,
	baseSnapshot domain.BaseSnapshot, localSnapshot, remoteSnapshot domain.Snapshot,
) (domain.SyncPlan, error) {

	if ctxErr := ctx.Err(); ctxErr != nil {
		return domain.SyncPlan{}, ctxErr
	}

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
		if ctxErr := ctx.Err(); ctxErr != nil {
			return domain.SyncPlan{}, ctxErr
		}
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
			s.skipSubtree(int(base[baseIdx].SubtreeSize), &baseIdx)
			s.skipSubtree(int(local[localIdx].SubtreeSize), &localIdx)
			s.skipSubtree(int(remote[remoteIdx].SubtreeSize), &remoteIdx)
			continue
		}

		localChange, remoteChange, conflict, potentialLocalChange, potentialRemoteChange := s.processEntry(
			baseEntry, localEntry, remoteEntry, baseSnapshot.IsInitial,
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

	appliedLocalChanges, appliedRemoteChanges, appliedConflicts, err := s.detectRenamesAndMoves(ctx, potentialLocalChanges, potentialRemoteChanges, base)
	if err != nil {
		return domain.SyncPlan{}, err
	}

	localChanges = append(localChanges, appliedLocalChanges...)
	remoteChanges = append(remoteChanges, appliedRemoteChanges...)
	conflicts = append(conflicts, appliedConflicts...)

	return domain.NewSyncPlan(localChanges, remoteChanges, conflicts), nil
}

func (s *ChangesPlannerWithTreeSkip) processEntry(
	baseEntry, localEntry, remoteEntry *domain.FileEntry, isInitial bool,
) (
	*domain.LocalChange,
	*domain.RemoteChange,
	*domain.Conflict,
	*potentialChange,
	*potentialChange,
) {
	switch {
	case s.shouldBeInitialMerge(baseEntry, localEntry, remoteEntry, isInitial):
		localChange, remoteChange := s.handleInitialMerge(localEntry, remoteEntry)
		return localChange, remoteChange, nil, nil, nil

	case s.isModified(baseEntry, localEntry, remoteEntry):
		localChange, remoteChange, conflict := s.handleModification(localEntry, remoteEntry)
		return localChange, remoteChange, conflict, nil, nil

	case s.isNewFile(baseEntry, localEntry, remoteEntry):
		localChange, remoteChange, potentialLocalChange, potentialRemoteChange, conflict := s.handleCreation(localEntry, remoteEntry)
		return localChange, remoteChange, conflict, potentialLocalChange, potentialRemoteChange

	case s.isDeleted(baseEntry, localEntry, remoteEntry):
		localChange, remoteChange, conflict, potentialLocalChange, potentialRemoteChange := s.handleDeletion(baseEntry, localEntry, remoteEntry)
		return localChange, remoteChange, conflict, potentialLocalChange, potentialRemoteChange
	}

	return nil, nil, nil, nil, nil
}

func (s *ChangesPlannerWithTreeSkip) handleDeletion(
	baseEntry, localEntry, remoteEntry *domain.FileEntry,
) (
	*domain.LocalChange,
	*domain.RemoteChange,
	*domain.Conflict,
	*potentialChange,
	*potentialChange,
) {

	if localEntry == nil && remoteEntry != nil {

		if remoteEntry.FileInfo.Metadata.IsDirectory {
			return nil, &domain.RemoteChange{
				OldRelativePath: remoteEntry.RelativePath,
				ChangeType:      domain.Delete,
			}, nil, nil, nil
		}

		// Конфликт удаления - удалено на local, но изменено на remote
		if remoteEntry.FileInfo.Hash != baseEntry.FileInfo.Hash {
			return nil, nil, &domain.Conflict{
				RemoteRelativePath: remoteEntry.RelativePath,
				BaseRelativePath:   baseEntry.RelativePath,
				RemoteModifiedUnix: remoteEntry.FileInfo.Metadata.ModifiedUnix,
				BaseModifiedUnix:   baseEntry.FileInfo.Metadata.ModifiedUnix,
				Conflict:           domain.ConflictLocalDeletedRemoteModified,
			}, nil, nil
		}

		// Удалено локально - нужно удалить на remote
		return nil, nil, nil, nil, &potentialChange{
			Changetype: domain.Delete,
			Hash:       baseEntry.FileInfo.Hash,
			Modify:     baseEntry.FileInfo.Metadata.ModifiedUnix,
			Path:       baseEntry.RelativePath,
		}
	}

	if remoteEntry == nil && localEntry != nil {
		if localEntry.FileInfo.Metadata.IsDirectory {
			return &domain.LocalChange{
				OldRelativePath: localEntry.RelativePath,
				ChangeType:      domain.Delete,
			}, nil, nil, nil, nil
		}
		// Конфликт удаления - удалено на remote, но изменено на local
		if localEntry.FileInfo.Hash != baseEntry.FileInfo.Hash {
			return nil, nil, &domain.Conflict{
				LocalRelativePath: localEntry.RelativePath,
				BaseRelativePath:  baseEntry.RelativePath,
				LocalModifiedUnix: localEntry.FileInfo.Metadata.ModifiedUnix,
				BaseModifiedUnix:  baseEntry.FileInfo.Metadata.ModifiedUnix,
				Conflict:          domain.ConflictRemoteDeletedLocalModified,
			}, nil, nil
		}

		// Удалено на remote - нужно удалить на local
		return nil, nil, nil, &potentialChange{
			Changetype: domain.Delete,
			Hash:       baseEntry.FileInfo.Hash,
			Modify:     baseEntry.FileInfo.Metadata.ModifiedUnix,
			Path:       baseEntry.RelativePath,
		}, nil
	}

	return nil, nil, nil, nil, nil
}

func (s *ChangesPlannerWithTreeSkip) handleCreation(
	localEntry, remoteEntry *domain.FileEntry,
) (*domain.LocalChange, *domain.RemoteChange, *potentialChange, *potentialChange, *domain.Conflict) {
	// создать на remote
	if localEntry != nil && remoteEntry == nil {
		if localEntry.FileInfo.Metadata.IsDirectory {
			return nil, &domain.RemoteChange{
				NewRelativePath: localEntry.RelativePath,
				ChangeType:      domain.CreateDir,
			}, nil, nil, nil
		}
		return nil, nil, nil, &potentialChange{
			Changetype: domain.CreateFile,
			Hash:       localEntry.FileInfo.Hash,
			Modify:     localEntry.FileInfo.Metadata.ModifiedUnix,
			Path:       localEntry.RelativePath,
		}, nil
	}

	// создать на local
	if remoteEntry != nil && localEntry == nil {
		if remoteEntry.FileInfo.Metadata.IsDirectory {
			return &domain.LocalChange{
				NewRelativePath: remoteEntry.RelativePath,
				ChangeType:      domain.CreateDir,
			}, nil, nil, nil, nil
		}
		return nil, nil, &potentialChange{
			Changetype: domain.CreateFile,
			Hash:       remoteEntry.FileInfo.Hash,
			Modify:     remoteEntry.FileInfo.Metadata.ModifiedUnix,
			Path:       remoteEntry.RelativePath,
		}, nil, nil
	}

	if localEntry.FileInfo.Metadata.IsDirectory && remoteEntry.FileInfo.Metadata.IsDirectory {
		return nil, nil, nil, nil, nil
	}

	if localEntry != nil && remoteEntry != nil {
		// Конфликт создания - оба создали файл с разным содержимым
		if localEntry.FileInfo.Hash != remoteEntry.FileInfo.Hash {
			return nil, nil, nil, nil, &domain.Conflict{
				LocalRelativePath:  localEntry.RelativePath,
				RemoteRelativePath: remoteEntry.RelativePath,
				LocalModifiedUnix:  localEntry.FileInfo.Metadata.ModifiedUnix,
				RemoteModifiedUnix: remoteEntry.FileInfo.Metadata.ModifiedUnix,
				Conflict:           domain.ConflictBothCreatedAtSamePathConflict,
			}
		}
	}
	return nil, nil, nil, nil, nil
}

func (s *ChangesPlannerWithTreeSkip) handleInitialMerge(
	localEntry, remoteEntry *domain.FileEntry,
) (*domain.LocalChange, *domain.RemoteChange) {
	if localEntry == nil && remoteEntry != nil {
		if remoteEntry.FileInfo.Metadata.IsDirectory {
			return &domain.LocalChange{
				NewRelativePath: remoteEntry.RelativePath,
				ChangeType:      domain.CreateDir,
			}, nil
		}

		return &domain.LocalChange{
			NewRelativePath: remoteEntry.RelativePath,
			ChangeType:      domain.CreateFile,
		}, nil
	}

	if localEntry != nil && remoteEntry == nil {
		if localEntry.FileInfo.Metadata.IsDirectory {
			return nil, &domain.RemoteChange{
				NewRelativePath: localEntry.RelativePath,
				ChangeType:      domain.CreateDir,
			}
		}

		return nil, &domain.RemoteChange{
			NewRelativePath: localEntry.RelativePath,
			ChangeType:      domain.CreateFile,
		}
	}

	if localEntry.FileInfo.Metadata.IsDirectory && remoteEntry.FileInfo.Metadata.IsDirectory {
		return nil, nil
	}

	if localEntry.FileInfo.Hash != remoteEntry.FileInfo.Hash {
		mergedName := domain.Path(fmt.Sprintf("%s-merged", localEntry.RelativePath))

		return &domain.LocalChange{
				NewRelativePath: mergedName,
				ChangeType:      domain.CreateFile,
			}, &domain.RemoteChange{
				NewRelativePath: mergedName,
				ChangeType:      domain.CreateFile,
			}
	}

	return nil, nil
}

func (s *ChangesPlannerWithTreeSkip) handleModification(
	local, remote *domain.FileEntry,
) (*domain.LocalChange, *domain.RemoteChange, *domain.Conflict) {
	if local.FileInfo.Metadata.IsDirectory && remote.FileInfo.Metadata.IsDirectory {
		return nil, nil, nil
	}

	if local.FileInfo.Metadata.ModifiedUnix > remote.FileInfo.Metadata.ModifiedUnix {
		return nil, &domain.RemoteChange{
			OldRelativePath: local.RelativePath,
			ChangeType:      domain.Modify,
		}, nil
	} else if local.FileInfo.Metadata.ModifiedUnix == remote.FileInfo.Metadata.ModifiedUnix {
		return nil, nil, &domain.Conflict{
			LocalRelativePath:  local.RelativePath,
			RemoteRelativePath: remote.RelativePath,
			LocalModifiedUnix:  local.FileInfo.Metadata.ModifiedUnix,
			RemoteModifiedUnix: remote.FileInfo.Metadata.ModifiedUnix,
			Conflict:           domain.ConflictBothModifiedAtSameTime,
		}
	}

	return &domain.LocalChange{
		OldRelativePath: remote.RelativePath,
		ChangeType:      domain.Modify,
	}, nil, nil
}

func (s *ChangesPlannerWithTreeSkip) detectRenamesAndMoves(
	ctx context.Context,
	potentialLocalChanges, potentialRemoteChanges []potentialChange,
	baseFiles []domain.FileEntry,
) ([]domain.LocalChange, []domain.RemoteChange, []domain.Conflict, error) {

	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, nil, nil, ctxErr
	}

	localChanges := make([]domain.LocalChange, 0, len(potentialLocalChanges))
	remoteChanges := make([]domain.RemoteChange, 0, len(potentialRemoteChanges))
	conflicts := make([]domain.Conflict, 0)

	localByHash := make(map[string][]potentialChange)
	remoteByHash := make(map[string][]potentialChange)

	for _, change := range potentialLocalChanges {
		localByHash[change.Hash] = append(localByHash[change.Hash], change)
	}
	for _, change := range potentialRemoteChanges {
		remoteByHash[change.Hash] = append(remoteByHash[change.Hash], change)
	}

	baseByPath := make(map[string]domain.FileEntry, len(baseFiles))
	for _, entry := range baseFiles {
		baseByPath[entry.RelativePath.String()] = entry
	}

	processedHashes := make(map[string]struct{})
	for h := range localByHash {
		processedHashes[h] = struct{}{}
	}
	for h := range remoteByHash {
		processedHashes[h] = struct{}{}
	}

	for hash := range processedHashes {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, nil, nil, ctxErr
		}

		localDeletes, localCreates := s.splitChanges(localByHash[hash])
		remoteDeletes, remoteCreates := s.splitChanges(remoteByHash[hash])

		// простой rename/move на local
		if len(localCreates) == 1 && len(localDeletes) == 1 &&
			len(remoteCreates) == 0 && len(remoteDeletes) == 0 {

			baseEntry, ok := s.findBaseEntryByPathAndHash(baseByPath, localDeletes[0].Path, hash)
			if !ok {
				continue
			}

			if localCreates[0].Path.String() == localDeletes[0].Path.String() {
				delete(localByHash, hash)
				continue
			}

			localChanges = append(localChanges,
				s.buildLocalRenameOrMove(*baseEntry, localCreates[0].Path),
			)
			delete(localByHash, hash)
			continue
		}

		// простой rename/move на remote
		if len(remoteCreates) == 1 && len(remoteDeletes) == 1 &&
			len(localCreates) == 0 && len(localDeletes) == 0 {

			baseEntry, ok := s.findBaseEntryByPathAndHash(baseByPath, remoteDeletes[0].Path, hash)
			if !ok {
				continue
			}

			if remoteCreates[0].Path.String() == remoteDeletes[0].Path.String() {
				delete(remoteByHash, hash)
				continue
			}

			remoteChanges = append(remoteChanges,
				s.buildRemoteRenameOrMove(*baseEntry, remoteCreates[0].Path),
			)
			delete(remoteByHash, hash)
			continue
		}

		// rename vs rename / move vs move
		if len(localCreates) == 1 && len(localDeletes) == 1 &&
			len(remoteCreates) == 1 && len(remoteDeletes) == 1 {

			localBaseEntry, ok := s.findBaseEntryByPathAndHash(baseByPath, localDeletes[0].Path, hash)
			if !ok {
				continue
			}

			remoteBaseEntry, ok := s.findBaseEntryByPathAndHash(baseByPath, remoteDeletes[0].Path, hash)
			if !ok {
				continue
			}

			// оба удаляют один и тот же base-путь
			if localDeletes[0].Path.String() == remoteDeletes[0].Path.String() {
				if localCreates[0].Path.String() == remoteCreates[0].Path.String() {
					delete(localByHash, hash)
					delete(remoteByHash, hash)
					continue
				}

				conflicts = append(conflicts,
					s.buildRenameMoveConflict(*localBaseEntry, &localCreates[0], &remoteCreates[0]),
				)
				delete(localByHash, hash)
				delete(remoteByHash, hash)
				continue
			}

			// разные base-пути — это два независимых очевидных move/rename
			if localCreates[0].Path.String() == remoteCreates[0].Path.String() {
				continue
			}

			localChanges = append(localChanges,
				s.buildLocalRenameOrMove(*localBaseEntry, localCreates[0].Path),
			)
			remoteChanges = append(remoteChanges,
				s.buildRemoteRenameOrMove(*remoteBaseEntry, remoteCreates[0].Path),
			)

			delete(localByHash, hash)
			delete(remoteByHash, hash)
			continue
		}
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, nil, nil, ctxErr
	}

	// оставшиеся изменения
	s.appendRemainingRemoteChanges(remoteByHash, &remoteChanges)
	s.appendRemainingLocalChanges(localByHash, &localChanges)

	return localChanges, remoteChanges, conflicts, nil
}

func (s *ChangesPlannerWithTreeSkip) findBaseEntryByPathAndHash(
	baseByPath map[string]domain.FileEntry,
	path domain.Path,
	hash string,
) (*domain.FileEntry, bool) {
	entry, ok := baseByPath[path.String()]
	if !ok {
		return nil, false
	}
	if entry.FileInfo.Hash != hash {
		return nil, false
	}
	return &entry, true
}

func (s *ChangesPlannerWithTreeSkip) splitChanges(
	changes []potentialChange,
) (deletes []potentialChange, creates []potentialChange) {

	for i := range changes {
		switch changes[i].Changetype {
		case domain.Delete:
			deletes = append(deletes, changes[i])
		case domain.CreateFile:
			creates = append(creates, changes[i])
		}
	}
	return
}

func (s *ChangesPlannerWithTreeSkip) buildLocalRenameOrMove(
	baseEntry domain.FileEntry,
	newPath domain.Path,
) domain.LocalChange {

	baseDir, baseName := filepath.Split(baseEntry.RelativePath.String())
	newDir, newName := filepath.Split(newPath.String())

	if baseDir == newDir && baseName != newName {
		return domain.LocalChange{
			OldRelativePath: baseEntry.RelativePath,
			NewRelativePath: newPath,
			ChangeType:      domain.Rename,
		}
	}

	return domain.LocalChange{
		OldRelativePath: baseEntry.RelativePath,
		NewRelativePath: newPath,
		ChangeType:      domain.Move,
	}
}
func (s *ChangesPlannerWithTreeSkip) buildRemoteRenameOrMove(
	baseEntry domain.FileEntry,
	newPath domain.Path,
) domain.RemoteChange {

	baseDir, baseName := filepath.Split(baseEntry.RelativePath.String())
	newDir, newName := filepath.Split(newPath.String())

	if baseDir == newDir && baseName != newName {
		return domain.RemoteChange{
			OldRelativePath: baseEntry.RelativePath,
			NewRelativePath: newPath,
			ChangeType:      domain.Rename,
		}
	}

	return domain.RemoteChange{
		OldRelativePath: baseEntry.RelativePath,
		NewRelativePath: newPath,
		ChangeType:      domain.Move,
	}
}

func (s *ChangesPlannerWithTreeSkip) buildRenameMoveConflict(
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

	// remoteCreate это изменение, которое нужно применить на remote следовательно переданный путь/время в нем это путь/время с local
	localRelativePath := remoteCreate.Path
	localModifiedUnix := remoteCreate.Modify
	// remoteCreate это изменение, которое нужно применить на local следовательно переданный путь/время  в нем это путь/время  с remote
	remoteRelativePath := localCreate.Path
	remoteModifiedUnix := localCreate.Modify

	return domain.Conflict{
		LocalRelativePath:  localRelativePath,
		RemoteRelativePath: remoteRelativePath,
		BaseRelativePath:   baseEntry.RelativePath,

		LocalModifiedUnix:  localModifiedUnix,
		RemoteModifiedUnix: remoteModifiedUnix,
		BaseModifiedUnix:   baseEntry.FileInfo.Metadata.ModifiedUnix,
		Conflict:           conflictType,
	}
}

func (s *ChangesPlannerWithTreeSkip) appendRemainingRemoteChanges(
	remoteByHash map[string][]potentialChange,
	target *[]domain.RemoteChange,
) {
	for _, changes := range remoteByHash {
		for _, change := range changes {
			switch change.Changetype {
			case domain.CreateFile:
				*target = append(*target, domain.RemoteChange{
					ChangeType:      domain.CreateFile,
					NewRelativePath: change.Path,
				})
			case domain.Delete:
				*target = append(*target, domain.RemoteChange{
					ChangeType:      domain.Delete,
					OldRelativePath: change.Path,
				})
			}
		}
	}
}

func (s *ChangesPlannerWithTreeSkip) appendRemainingLocalChanges(
	localByHash map[string][]potentialChange,
	target *[]domain.LocalChange,
) {
	for _, changes := range localByHash {
		for _, change := range changes {
			switch change.Changetype {
			case domain.CreateFile:
				*target = append(*target, domain.LocalChange{
					ChangeType:      domain.CreateFile,
					NewRelativePath: change.Path,
				})
			case domain.Delete:
				*target = append(*target, domain.LocalChange{
					ChangeType:      domain.Delete,
					OldRelativePath: change.Path,
				})
			}
		}
	}
}

func (s *ChangesPlannerWithTreeSkip) isRenamedRenamedConflict(
	baseEntryDir, localDir, baseEntryFileName, localFileName, remoteDir, remoteFileName string,
) bool {
	return baseEntryDir == localDir && baseEntryFileName != localFileName &&
		baseEntryDir == remoteDir && baseEntryFileName != remoteFileName
}

func (s *ChangesPlannerWithTreeSkip) isMovedMovedConflict(
	baseEntryDir, localDir, remoteDir string,
) bool {
	return baseEntryDir != localDir &&
		baseEntryDir != remoteDir &&
		localDir != remoteDir
}

func (s *ChangesPlannerWithTreeSkip) pathAt(entries []domain.FileEntry, idx int) string {
	if idx >= len(entries) {
		return sentinel
	}
	return entries[idx].RelativePath.String()
}

func (s *ChangesPlannerWithTreeSkip) minString(a, b, c string) string {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

func (s *ChangesPlannerWithTreeSkip) canSkipSubtreeAt(b, l, r *domain.FileEntry, currentPath string) bool {
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

func (s *ChangesPlannerWithTreeSkip) skipSubtree(subtreeSize int, idx *int) {
	*idx += subtreeSize
}

func (s *ChangesPlannerWithTreeSkip) isDeleted(base, local, remote *domain.FileEntry) bool {
	return base != nil && (local == nil || remote == nil)
}

func (s *ChangesPlannerWithTreeSkip) isNewFile(base, local, remote *domain.FileEntry) bool {
	if base != nil {
		return false
	}
	return local != nil || remote != nil
}

func (s *ChangesPlannerWithTreeSkip) isModified(base, local, remote *domain.FileEntry) bool {
	return base != nil && local != nil && remote != nil && local.FileInfo.Hash != remote.FileInfo.Hash
}

func (s *ChangesPlannerWithTreeSkip) shouldBeInitialMerge(base, local, remote *domain.FileEntry, isInitial bool) bool {
	return isInitial && (s.isDeleted(base, local, remote) || s.isNewFile(base, local, remote) || s.isModified(base, local, remote))
}
