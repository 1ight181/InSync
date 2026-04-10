package models

import (
	"strings"
)

type MDnsBrowserConfig struct {
	ServerServiceType string `mapstructure:"server_service_type"`
	ServerDomain      string `mapstructure:"server_domain"`
	Interfaces        string `mapstructure:"interfaces"`
}

func (mc *MDnsBrowserConfig) Validate() error {
	if mc.ServerServiceType == "" {
		return ErrMDnsBrowserServerServiceTypeIsEmpty
	}
	if mc.ServerDomain == "" {
		return ErrMDnsBrowserServerDomainIsEmpty
	}

	return nil
}

func (mc *MDnsBrowserConfig) GetInterfaces() []string {
	return strings.Split(mc.Interfaces, ";")
}
