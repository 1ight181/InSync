package models

import (
	"strings"
)

type MDnsServerConfig struct {
	ServiceType string `mapstructure:"service_type"`
	Domain      string `mapstructure:"domain"`
	Port        int    `mapstructure:"port"`
	Interfaces  string `mapstructure:"interfaces"`
}

func (mc *MDnsServerConfig) Validate() error {
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
