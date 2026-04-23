package connection

import "insync/internal/interfaces"

type IClientHolder interface {
	SetClient(client interfaces.IClient)
	CurrentClient() (interfaces.IClient, error)
	ReleaseClient()
}
