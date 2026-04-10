package models

import (
	"strings"
)

type MDnsBrowserConfig struct {
	ServerServiceType         string `mapstructure:"server_service_type"`
	ServerInstanceNamePostfix string `mapstructure:"server_instance_name_postfix"`
	ServerDomain              string `mapstructure:"server_domain"`
	Interfaces                string `mapstructure:"interfaces"`
}

func (mc *MDnsBrowserConfig) Validate() error {
	if mc.ServerServiceType == "" {
		return ErrMDnsBrowserServerServiceTypeIsEmpty
	}
	if mc.ServerInstanceNamePostfix == "" {
		return ErrMDnsBrowserServerInstanceNamePostfixIsEmpty
	}
	if mc.ServerDomain == "" {
		return ErrMDnsBrowserServerDomainIsEmpty
	}

	return nil
}

func (mc *MDnsBrowserConfig) GetInterfaces() []string {
	return strings.Split(mc.Interfaces, ";")
}
