package domain

import "errors"

var (
	ErrInvalidDeviceId = errors.New("deviceId не может быть пустым")
)

type DeviceId string

func (d DeviceId) String() string {
	return string(d)
}

func NewDeviceId(deviceId string) (DeviceId, error) {
	if deviceId == "" {
		return "", ErrInvalidDeviceId
	}
	return DeviceId(deviceId), nil
}
