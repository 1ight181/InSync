package sync

import (
	"context"
	"insync/internal/domain"
)

type IPlanResolver interface {
	Resolve(ctx context.Context, rootName domain.RootName, shouldUseCache bool) (domain.SyncPlan, error)
}
