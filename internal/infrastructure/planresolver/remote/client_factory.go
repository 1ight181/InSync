package remote

import "insync/internal/interfaces"

type IClientFactory interface {
	CurrentClient() interfaces.IClient
}
