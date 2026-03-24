package models

import (
	"fmt"
	"strings"
)

type MdnsServerConfig struct {
	InstanceName string `mapstructure:"instance_name"`
	ServiceType  string `mapstructure:"service_type"`
	Domain       string `mapstructure:"domain"`
	Port         int    `mapstructure:"port"`
	Interfaces   string `mapstructure:"interfaces"`
}

func (mc *MdnsServerConfig) Validate() error {
	if mc.InstanceName == "" {
		return ErrMdnsInstanceNameIsEmpty
	}
	if mc.ServiceType == "" {
		return ErrMdnsServiceTypeIsEmpty
	}
	if mc.Domain == "" {
		return ErrMdnsDomainIsEmpty
	}
	if mc.Port < 0 || mc.Port > 65535 {
		return ErrMdnsPortIsInvalid
	}

	return nil
}

func (mc *MdnsServerConfig) GetInterfaces() []string {
	return strings.Split(mc.Interfaces, ";")
}

func (mc *MdnsServerConfig) GetFullServerName() string {
	return fmt.Sprintf("%s.%s", mc.InstanceName, mc.Domain)
}
