package local

import "context"

type ILocalDeviceIdCreator interface {
	CreateLocalDeviceId(ctx context.Context) (string, error)
}
