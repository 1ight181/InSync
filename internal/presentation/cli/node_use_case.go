package cli

import "insync/internal/domain"

type INodeUseCase interface {
	ShowNodeNames() (nodeNamesChan chan domain.NodeName, err error)
}
