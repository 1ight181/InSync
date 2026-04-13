package scan

import "insync/internal/domain"

type IChangesScanner interface {
	Scan(localSnapshot, remoteSnapshot []domain.FileEntry) ([]domain.SyncChange, error)
}
