package scan

import "insync/internal/domain"

type ChangesScanner struct {
	rootResolver IRootResolver
}

func NewChangesScanner() *ChangesScanner {
	return &ChangesScanner{}
}

func (s *ChangesScanner) Scan(rootName string) domain.SyncChange {
	return domain.SyncChange{}
}
