package cli

import (
	"context"
	"insync/internal/domain"
)

type IInitUseCase interface {
	InitFromLocal(ctx context.Context, rootName domain.RootName) error
	InitFromRemote(ctx context.Context, rootName domain.RootName) error
}
