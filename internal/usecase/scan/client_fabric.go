package scan

import "insync/internal/interfaces"

type IClientFabric interface {
	CurrentClient() interfaces.IClient
}
