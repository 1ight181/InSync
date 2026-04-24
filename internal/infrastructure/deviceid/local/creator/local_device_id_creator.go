package creator

import (
	"bytes"
	"errors"
	"insync/internal/domain"
	"io"
	"io/fs"

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
	return &LocalDeviceIdCreator{
		filsSys:          opts.FileSys,
		deviceIdFilePath: opts.DeviceIdFilePath,
	}, nil
}

func (c *LocalDeviceIdCreator) ReadOrCreateLocalDeviceId() (domain.DeviceId, error) {
	deviceIdFile, err := c.filsSys.Open(c.deviceIdFilePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return c.generateAndWriteDeviceId()
		}
		return "", err
	}

	deviceIdRaw, err := io.ReadAll(deviceIdFile)
	if err != nil {
		return "", err
	}

	deviceIdFile.Close()

	if len(deviceIdRaw) == 0 {
		return c.generateAndWriteDeviceId()
	}

	return domain.NewDeviceId(string(deviceIdRaw))

}

func (c *LocalDeviceIdCreator) generateAndWriteDeviceId() (domain.DeviceId, error) {
	deviceId := c.generateDeviceId()
	if err := c.writeOrCreateDeviceId(deviceId); err != nil {
		return "", err
	}
	return deviceId, nil
}

func (c *LocalDeviceIdCreator) generateDeviceId() domain.DeviceId {
	return domain.DeviceId(uuid.New().String())
}

func (c LocalDeviceIdCreator) writeOrCreateDeviceId(deviceId domain.DeviceId) error {
	reader := bytes.NewReader([]byte(deviceId.String()))

	err := c.filsSys.AtomicWrite(c.deviceIdFilePath, reader)
	if err != nil {
		return err
	}

	return nil
}
