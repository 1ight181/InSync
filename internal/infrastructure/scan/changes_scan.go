package scan

import "insync/internal/domain"

type ChangesScanner struct {
}

func NewChangesScanner() *ChangesScanner {
	return &ChangesScanner{}
}

func (s *ChangesScanner) Scan(localSnapshot, remoteSnapshot []domain.FileEntry) domain.SyncChange {
	return domain.SyncChange{}
}
