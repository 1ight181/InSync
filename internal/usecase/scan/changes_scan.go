package scan

import "insync/internal/domain"

type IChangesScanner interface {
	Scan(localFileEntries, remoteFileEntries []domain.FileEntry) ([]domain.SyncChange, error)
}
