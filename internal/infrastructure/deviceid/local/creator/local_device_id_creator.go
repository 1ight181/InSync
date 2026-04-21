package creator

import (
	"bytes"
	"errors"
	"insync/internal/domain"

	"github.com/google/uuid"
)

type LocalDeviceIdCreator struct {
	filsSys          IFileSystem
	deviceIdFilePath domain.Path
}

type LocalDeviceIdCreatorOptions struct {
	FileSys          IFileSystem
	DeviceIdFilePath domain.Path
}

var (
	ErrInvalidOpts = errors.New("Все поля LocalDeviceIdCreatorOptions должны быть заполнены")
)

func NewLocalDeviceIdCreator(opts LocalDeviceIdCreatorOptions) (*LocalDeviceIdCreator, error) {
	if opts.FileSys == nil || opts.DeviceIdFilePath == "" {
		return nil, ErrInvalidOpts
	}
	return &LocalDeviceIdCreator{filsSys: opts.FileSys}, nil
}

func (c *LocalDeviceIdCreator) CreateLocalDeviceId() (domain.DeviceId, error) {
	uuid := c.generateDeviceId()

	reader := bytes.NewReader([]byte(uuid))

	err := c.filsSys.AtomicWrite(c.deviceIdFilePath, reader)
	if err != nil {
		return "", err
	}
	deviceId, err := domain.NewDeviceId(uuid)
	if err != nil {
		return "", err
	}

	return deviceId, nil
}

func (c *LocalDeviceIdCreator) generateDeviceId() string {
	return uuid.New().String()
}
