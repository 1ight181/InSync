package models

import (
	conferr "insync/internal/config/errors"
	"strconv"
)

type ClientConfig struct {
	ClientIp          string
	ClientPort        int
	ClientNetworkType string
}

func (cc *ClientConfig) Validate() error {
	if cc.ClientIp == "" {
		return conferr.ErrClientIpIsEmpty
	}
	if cc.ClientPort < 0 || cc.ClientPort > 65535 {
		return conferr.ErrClientPortIsInvalid
	}
	if cc.ClientNetworkType == "" {
		return conferr.ErrClientNetworkTypeIsEmpty
	}

	return nil
}

func (cc *ClientConfig) GetClientAddress() string {
	return cc.ClientIp + ":" + strconv.Itoa(sc.ClientPort)
}
