package cli

import (
	"context"
	"insync/internal/domain"
)

type INodeUseCase interface {
	ShowNodeNames(ctx context.Context) (nodeNamesChan chan domain.NodeNameWithAlias, err error)
}
