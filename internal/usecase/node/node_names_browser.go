package connect

import "insync/internal/domain"

type INodeNamesBrowser interface {
	BrowseNodeNames() (nodeNamesChan chan domain.NodeName, err error)
}
