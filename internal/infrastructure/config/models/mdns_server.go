package models

import (
	"fmt"
	"strings"
)

type MDnsServerConfig struct {
	InstanceName        string `mapstructure:"instance_name"`
	InstanceNamePostfix string `mapstructure:"instance_name_postfix"`
	ServiceType         string `mapstructure:"service_type"`
	Domain              string `mapstructure:"domain"`
	Port                int    `mapstructure:"port"`
	Interfaces          string `mapstructure:"interfaces"`
}

func (mc *MDnsServerConfig) Validate() error {
	if mc.InstanceName == "" {
		return ErrMDnsServerServiceTypeIsEmpty
	}
	if mc.InstanceNamePostfix == "" {
		return ErrMDnsServerInstanceNamePostfixIsEmpty
	}
	if mc.ServiceType == "" {
		return ErrMDnsServerServiceTypeIsEmpty
	}
	if mc.Domain == "" {
		return ErrMDnsServerDomainIsEmpty
	}
	if mc.Port < 0 || mc.Port > 65535 {
		return ErrMDnsServerPortIsInvalid
	}

	return nil
}

func (mc *MDnsServerConfig) GetInterfaces() []string {
	return strings.Split(mc.Interfaces, ";")
}

func (mc *MDnsServerConfig) GetFullServerName() string {
	return fmt.Sprintf("%s.%s", mc.InstanceName, mc.Domain)
}

func (mc *MDnsServerConfig) GetFullInstanceName() string {
	return fmt.Sprintf("%s.%s", mc.InstanceName, mc.InstanceNamePostfix)
}
