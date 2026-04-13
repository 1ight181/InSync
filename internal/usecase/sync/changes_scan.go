package sync

import "insync/internal/domain"

type IChangesScanner interface {
	Scan(rootName domain.RootName) ([]domain.SyncChange, error)
}
