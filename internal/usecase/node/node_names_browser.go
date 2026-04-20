package connect

import (
	"context"
	"insync/internal/domain"
)

type INodeNamesBrowser interface {
	BrowseNodeNames(ctx context.Context) (nodeNamesChan chan domain.NodeName, err error)
}
