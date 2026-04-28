package base

import "insync/internal/interfaces"

type IClientFactory interface {
	CurrentClient() (interfaces.IClient, error)
}
