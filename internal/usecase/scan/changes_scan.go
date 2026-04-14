package scan

import "insync/internal/domain"

type IChangesScanner interface {
	Scan(localFileEntries, remoteFileEntries domain.Snapshot) ([]domain.SyncChange, error)
}
