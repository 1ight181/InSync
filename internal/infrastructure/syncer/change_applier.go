package syncer

import (
	"context"
	"insync/internal/domain"
)

type IChangeApplier interface {
	ApplyLocal(ctx context.Context, rootName domain.RootName, change domain.LocalChange) error
	ApplyRemote(ctx context.Context, rootName domain.RootName, change domain.RemoteChange) error
}
