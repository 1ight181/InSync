package cli

import (
	"context"
	"insync/internal/domain"
)

type IInitUseCase interface {
	Init(ctx context.Context, rootName domain.RootName) error
}
