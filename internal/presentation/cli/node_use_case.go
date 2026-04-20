package cli

import (
	"context"
	"insync/internal/domain"
)

type INodeUseCase interface {
	ShowNodeNames(ctx context.Context) (nodeNamesChan chan domain.NodeName, err error)
}
