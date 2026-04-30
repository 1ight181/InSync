package local

import (
	"errors"
	"insync/internal/domain"
)

type LocalDeviceIdResolver struct {
	localDeviceIdCreator ILocalDeviceIdCreator
	currentDeviceId      domain.DeviceId
}

var (
	ErrInvalidOpts = errors.New("Все поля LocalDeviceIdResolverOptions должны быть заполнены")
)

func NewLocalDeviceIdResolver(localDeviceIdCreator ILocalDeviceIdCreator) (*LocalDeviceIdResolver, error) {
	if localDeviceIdCreator == nil {
		return nil, ErrInvalidOpts
	}
	return &LocalDeviceIdResolver{localDeviceIdCreator: localDeviceIdCreator}, nil
}

func (r *LocalDeviceIdResolver) Resolve() (domain.DeviceId, error) {
	if r.currentDeviceId == "" {
		newDeviceId, err := r.localDeviceIdCreator.ReadOrCreateLocalDeviceId()
		if err != nil {
			return "", err
		}

		r.currentDeviceId = newDeviceId
	}

	return r.currentDeviceId, nil
}
