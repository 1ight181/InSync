package creator

import (
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

func NewLocalDeviceIdCreator(opts LocalDeviceIdCreatorOptions) *LocalDeviceIdCreator {
	if opts.FileSys == nil || opts.DeviceIdFilePath == "" {
		panic("Все поля LocalDeviceIdCreatorOptions должны быть заполнены")
	}
	return &LocalDeviceIdCreator{filsSys: opts.FileSys}
}

func (c *LocalDeviceIdCreator) CreateDeviceId() (string, error) {
	file, err := c.filsSys.Create(c.deviceIdFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	uuid := c.generateDeviceId()

	file.Write([]byte(uuid))

	return uuid, nil
}

func (c *LocalDeviceIdCreator) generateDeviceId() string {
	return uuid.New().String()
}
