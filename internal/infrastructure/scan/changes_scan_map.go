package scan

// import (
// 	"context"
// 	"insync/internal/domain"
// 	"log/slog"
// 	"path/filepath"
// )

// type sideChange struct {
// 	isLocal      bool
// 	localChange  domain.LocalChange
// 	remoteChange domain.RemoteChange
// }

// type ChangesScannerWithMap struct {
// 	logger    *slog.Logger
// 	loggerCtx context.Context
// }

// type ChangesScannerWithMapOptions struct {
// 	Logger *slog.Logger
// }

// func NewChangesScannerWithMap(opts ChangesScannerWithMapOptions) *ChangesScannerWithMap {
// 	if opts.Logger == nil {
// 		panic("Logger must be provided to ChangesScannerWithMap")
// 	}
// 	return &ChangesScannerWithMap{
// 		logger:    opts.Logger,
// 		loggerCtx: context.Background(),
// 	}
// }

// func (s *ChangesScannerWithMap) Scan(
// 	baseSnapshot, localSnapshot, remoteSnapshot domain.Snapshot,
// ) ([]domain.LocalChange, []domain.RemoteChange) {

// 	var localChanges []domain.LocalChange
// 	var remoteChanges []domain.RemoteChange

// 	baseMap := s.createPathMap(baseSnapshot.Files)
// 	localMap := s.createPathMap(localSnapshot.Files)
// 	remoteMap := s.createPathMap(remoteSnapshot.Files)

// 	baseByHash := s.createHashMap(baseSnapshot.Files)
// 	localByHash := s.createHashMap(localSnapshot.Files)
// 	remoteByHash := s.createHashMap(remoteSnapshot.Files)

// 	allPaths := s.unionPaths(baseMap, localMap, remoteMap)

// 	for pathStr := range allPaths {
// 		baseEntry := baseMap[pathStr]
// 		localEntry := localMap[pathStr]
// 		remoteEntry := remoteMap[pathStr]

// 		switch {
// 		case s.isDeleted(baseEntry, localEntry, remoteEntry):
// 			if ch := s.handleDeletion(baseEntry, localEntry, remoteEntry); ch != nil {
// 				if ch.isLocal {
// 					localChanges = append(localChanges, ch.localChange)
// 				} else {
// 					remoteChanges = append(remoteChanges, ch.remoteChange)
// 				}
// 			}

// 		case s.isNewFile(baseEntry, localEntry, remoteEntry):
// 			if ch := s.handleCreation(localEntry, remoteEntry); ch != nil {
// 				if ch.isLocal {
// 					localChanges = append(localChanges, ch.localChange)
// 				} else {
// 					remoteChanges = append(remoteChanges, ch.remoteChange)
// 				}
// 			}

// 		case s.isModified(baseEntry, localEntry, remoteEntry):
// 			if ch := s.handleModification(baseEntry, localEntry, remoteEntry); ch != nil {
// 				if ch.isLocal {
// 					localChanges = append(localChanges, ch.localChange)
// 				} else {
// 					remoteChanges = append(remoteChanges, ch.remoteChange)
// 				}
// 			}

// 		}
// 	}

// 	s.detectRenamesAndMoves(baseByHash, localByHash, remoteByHash, &localChanges, &remoteChanges)

// 	return localChanges, remoteChanges
// }

// func (s *ChangesScannerWithMap) createPathMap(entries []domain.FileEntry) map[string]*domain.FileEntry {
// 	m := make(map[string]*domain.FileEntry, len(entries))
// 	for i := range entries {
// 		m[entries[i].RelativePath.String()] = &entries[i]
// 	}
// 	return m
// }

// func (s *ChangesScannerWithMap) unionPaths(maps ...map[string]*domain.FileEntry) map[string]struct{} {
// 	all := make(map[string]struct{})
// 	for _, m := range maps {
// 		for p := range m {
// 			all[p] = struct{}{}
// 		}
// 	}
// 	return all
// }

// func (s *ChangesScannerWithMap) isDeleted(base, local, remote *domain.FileEntry) bool {
// 	return base != nil && local == nil && remote == nil
// }

// func (s *ChangesScannerWithMap) isNewFile(base, local, remote *domain.FileEntry) bool {
// 	if base != nil {
// 		return false
// 	}
// 	return (local != nil) != (remote != nil) //  XOR
// }
// func (s *ChangesScannerWithMap) isModified(base, local, remote *domain.FileEntry) bool {
// 	if local == nil || remote == nil {
// 		return false
// 	}
// 	// Если пути совпадают, но содержимое изменилось
// 	return local.RelativePath.String() == remote.RelativePath.String() &&
// 		local.FileInfo.Hash != remote.FileInfo.Hash
// }

// func (s *ChangesScannerWithMap) handleModification(base, local, remote *domain.FileEntry) *sideChange {
// 	if local.FileInfo.Metadata.ModifiedUnix >= remote.FileInfo.Metadata.ModifiedUnix {
// 		return &sideChange{
// 			isLocal: false,
// 			remoteChange: domain.RemoteChange{
// 				OldRelativePath: local.RelativePath,
// 				ChangeType:      domain.Modify,
// 				ModifiedUnix:    local.FileInfo.Metadata.ModifiedUnix,
// 			},
// 		}
// 	}
// 	return &sideChange{
// 		isLocal: true,
// 		localChange: domain.LocalChange{
// 			OldRelativePath: remote.RelativePath,
// 			ChangeType:      domain.Modify,
// 			ModifiedUnix:    remote.FileInfo.Metadata.ModifiedUnix,
// 		},
// 	}
// }

// func (s *ChangesScannerWithMap) detectRenamesAndMoves(
// 	baseByHash, localByHash, remoteByHash map[string]*domain.FileEntry,
// 	localChanges *[]domain.LocalChange,
// 	remoteChanges *[]domain.RemoteChange,
// ) {
// 	seen := make(map[string]bool)

// 	for hash, _ := range baseByHash {
// 		localEntry := localByHash[hash]
// 		remoteEntry := remoteByHash[hash]

// 		if localEntry == nil || remoteEntry == nil {
// 			continue
// 		}

// 		if seen[hash] {
// 			continue
// 		}
// 		seen[hash] = true

// 		changeType := domain.Rename
// 		if filepath.Dir(localEntry.RelativePath.String()) != filepath.Dir(remoteEntry.RelativePath.String()) {
// 			changeType = domain.Move
// 		}

// 		if localEntry.FileInfo.Metadata.ModifiedUnix >= remoteEntry.FileInfo.Metadata.ModifiedUnix {

// 			*remoteChanges = append(*remoteChanges, domain.RemoteChange{
// 				OldRelativePath: remoteEntry.RelativePath,
// 				NewRelativePath: localEntry.RelativePath,
// 				ChangeType:      changeType,
// 				ModifiedUnix:    localEntry.FileInfo.Metadata.ModifiedUnix,
// 			})
// 		} else {
// 			// remote "победил"
// 			*localChanges = append(*localChanges, domain.LocalChange{
// 				OldRelativePath: localEntry.RelativePath,
// 				NewRelativePath: remoteEntry.RelativePath,
// 				ChangeType:      changeType,
// 				ModifiedUnix:    remoteEntry.FileInfo.Metadata.ModifiedUnix,
// 			})
// 		}
// 	}
// }

// // handleDeletion — удаление файла на одной из сторон
// func (s *ChangesScannerWithMap) handleDeletion(base, local, remote *domain.FileEntry) *sideChange {
// 	if local == nil && remote != nil {
// 		// Удалено на local
// 		if remote.FileInfo.Metadata.ModifiedUnix > base.FileInfo.Metadata.ModifiedUnix {
// 			return &sideChange{isLocal: true, localChange: domain.LocalChange{
// 				OldRelativePath: remote.RelativePath,
// 				ChangeType:      domain.Delete,
// 				ModifiedUnix:    remote.FileInfo.Metadata.ModifiedUnix,
// 			}}
// 		}
// 		return &sideChange{isLocal: false, remoteChange: domain.RemoteChange{
// 			OldRelativePath: base.RelativePath,
// 			ChangeType:      domain.Delete,
// 			ModifiedUnix:    base.FileInfo.Metadata.ModifiedUnix,
// 		}}
// 	}

// 	if remote == nil && local != nil {
// 		// Удалено на remote
// 		if local.FileInfo.Metadata.ModifiedUnix > base.FileInfo.Metadata.ModifiedUnix {
// 			return &sideChange{isLocal: false, remoteChange: domain.RemoteChange{
// 				OldRelativePath: local.RelativePath,
// 				ChangeType:      domain.Delete,
// 				ModifiedUnix:    local.FileInfo.Metadata.ModifiedUnix,
// 			}}
// 		}
// 		return &sideChange{isLocal: true, localChange: domain.LocalChange{
// 			OldRelativePath: base.RelativePath,
// 			ChangeType:      domain.Delete,
// 			ModifiedUnix:    base.FileInfo.Metadata.ModifiedUnix,
// 		}}
// 	}
// 	return nil
// }

// func (s *ChangesScannerWithMap) handleCreation(local, remote *domain.FileEntry) *sideChange {
// 	if local != nil && remote == nil {
// 		return &sideChange{isLocal: false, remoteChange: domain.RemoteChange{
// 			NewRelativePath: local.RelativePath,
// 			ChangeType:      domain.Create,
// 			ModifiedUnix:    local.FileInfo.Metadata.ModifiedUnix,
// 		}}
// 	}
// 	if remote != nil && local == nil {
// 		return &sideChange{isLocal: true, localChange: domain.LocalChange{
// 			NewRelativePath: remote.RelativePath,
// 			ChangeType:      domain.Create,
// 			ModifiedUnix:    remote.FileInfo.Metadata.ModifiedUnix,
// 		}}
// 	}
// 	return nil
// }

// func (s *ChangesScannerWithMap) createHashMap(entries []domain.FileEntry) map[string]*domain.FileEntry {
// 	m := make(map[string]*domain.FileEntry, len(entries))
// 	for i := range entries {
// 		if entries[i].FileInfo.Hash != "" {
// 			m[entries[i].FileInfo.Hash] = &entries[i]
// 		}
// 	}
// 	return m
// }
