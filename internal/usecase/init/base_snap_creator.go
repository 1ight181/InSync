package init

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotCreator interface {
	CreateBaseSnapshot(ctx context.Context,
		baseSnapshot domain.Snapshot,
		rootName domain.RootName,
	) error
}
