package models

import "path/filepath"

type DeviceIdResolverConfig struct {
	DeviceIdDir      string `mapstructure:"device_id_dir"`
	DeviceIdFileName string `mapstructure:"device_id_file_name"`
}

func (mc *DeviceIdResolverConfig) Validate() error {
	if mc.DeviceIdDir == "" {
		return ErrMDnsServerInstanceNamePostfixIsEmpty
	}

	if mc.DeviceIdFileName == "" {
		return ErrMDnsServerInstanceNamePostfixIsEmpty
	}

	return nil
}

func (mc *DeviceIdResolverConfig) GetDeviceIdFilePath() string {
	return filepath.Join(mc.DeviceIdDir, mc.DeviceIdFileName)
}
