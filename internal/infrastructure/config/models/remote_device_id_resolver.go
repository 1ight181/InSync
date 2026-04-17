package models

type RemoteDeviceIdResolverConfig struct {
	MDnsServerInstanceNamePostfix string `mapstructure:"mdns_server_instance_name_postfix"`
}

func (mc *RemoteDeviceIdResolverConfig) Validate() error {
	if mc.MDnsServerInstanceNamePostfix == "" {
		return ErrMDnsServerInstanceNamePostfixIsEmpty
	}

	return nil
}
