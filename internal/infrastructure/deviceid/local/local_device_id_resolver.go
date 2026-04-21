package local

import (
	"errors"
	"insync/internal/domain"
)

type LocalDeviceIdResolver struct {
	localDeviceIdCreator ILocalDeviceIdCreator
	currentDeviceId      domain.DeviceId
}

type LocalDeviceIdResolverOptions struct {
	LocalDeviceIdCreator ILocalDeviceIdCreator
}

var (
	ErrInvalidOpts = errors.New("Все поля LocalDeviceIdResolverOptions должны быть заполнены")
)

func NewLocalDeviceIdResolver(opts LocalDeviceIdResolverOptions) (*LocalDeviceIdResolver, error) {
	if opts.LocalDeviceIdCreator == nil {
		return nil, ErrInvalidOpts
	}
	return &LocalDeviceIdResolver{localDeviceIdCreator: opts.LocalDeviceIdCreator}, nil
}

func (r *LocalDeviceIdResolver) Resolve() (domain.DeviceId, error) {
	if r.currentDeviceId == "" {
		newDeviceId, err := r.localDeviceIdCreator.CreateLocalDeviceId()
		if err != nil {
			return "", err
		}

		r.currentDeviceId = newDeviceId
	}

	return r.currentDeviceId, nil
}
