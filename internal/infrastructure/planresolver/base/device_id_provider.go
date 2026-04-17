package base

type IDeviceIdProvider interface {
	GetCurrentLocalDeviceId() string
	GetCurrentRemoteDeviceId() (string, error)
}
