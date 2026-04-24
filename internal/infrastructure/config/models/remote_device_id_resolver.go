package models

type RemoteDeviceIdResolverConfig struct {
	MDnsServerInstanceNamePostfix string `mapstructure:"mdns_server_instance_name_postfix"`
	MDnsServerDomain              string `mapstructure:"mdns_server_domain"`
}

func (mc *RemoteDeviceIdResolverConfig) Validate() error {
	if mc.MDnsServerInstanceNamePostfix == "" {
		return ErrMDnsServerInstanceNamePostfixIsEmpty
	}

	if mc.MDnsServerDomain == "" {
		return ErrMDnsServerDomainIsEmpty
	}

	return nil
}
