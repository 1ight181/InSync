package scan

import "insync/internal/domain"

type ChangesScanner struct {
}

func NewChangesScanner() *ChangesScanner {
	return &ChangesScanner{}
}

func (s *ChangesScanner) Scan(localFileEntries, remoteFileEntries []domain.FileEntry) []domain.SyncChange {
	return nil
}
