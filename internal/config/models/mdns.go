package models

import (
	conferr "insync/internal/config/errors"
	"strings"
)

type MdnsConfig struct {
	SelfInstanceName   string `mapstructure:"self_instance_name"`
	InstanceNameToSync string `mapstructure:"instance_name_to_sync"`
	ServiceType        string `mapstructure:"service_type"`
	Domain             string `mapstructure:"domain"`
	Port               int    `mapstructure:"port"`
	Interfaces         string
}

func (mc *MdnsConfig) Validate() error {
	if mc.SelfInstanceName == "" {
		return conferr.ErrMdnsSelfInstanceNameIsEmpty
	}
	if mc.InstanceNameToSync == "" {
		return conferr.ErrMdnsInstanceNameToSyncIsEmpty
	}
	if mc.ServiceType == "" {
		return conferr.ErrMdnsServiceTypeIsEmpty
	}
	if mc.Domain == "" {
		return conferr.ErrMdnsDomainIsEmpty
	}
	if mc.Port < 0 || mc.Port > 65535 {
		return conferr.ErrMdnsPortIsInvalid
	}

	interfaces := mc.GetInterfaces()
	if len(interfaces) == 1 && interfaces[0] == "" {
		return conferr.ErrMdnsInterfacesIsEmpty
	}

	return nil
}

func (mc *MdnsConfig) GetInterfaces() []string {
	return strings.Split(mc.Interfaces, ";")
}
