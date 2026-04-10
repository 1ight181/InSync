package models

import (
	"strings"
)

type MDnsBrowserConfig struct {
	ServiceType string `mapstructure:"service_type"`
	Domain      string `mapstructure:"domain"`
	Interfaces  string `mapstructure:"interfaces"`
}

func (mc *MDnsBrowserConfig) Validate() error {
	if mc.ServiceType == "" {
		return ErrMDnsBrowserServiceTypeIsEmpty
	}
	if mc.Domain == "" {
		return ErrMDnsBrowserDomainIsEmpty
	}

	return nil
}

func (mc *MDnsBrowserConfig) GetInterfaces() []string {
	return strings.Split(mc.Interfaces, ";")
}
