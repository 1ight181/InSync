package syncer

import (
	"insync/internal/domain"
)

type IConflictResolver interface {
	Resolve(conflicts domain.Conflict, userDecision domain.Decision) (requiredChange domain.SyncChange, isLocal bool, err error)
}
