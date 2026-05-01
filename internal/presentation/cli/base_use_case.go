package cli

import (
	"context"
	"insync/internal/domain"
)

type IBaseUseCase interface {
	DeleteBaseSnapshots(ctx context.Context, rootName domain.RootName) error
}
