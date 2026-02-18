package models

import (
	conferr "insync/internal/config/errors"
	"strconv"
)

// Реализует интерфейс ConfigModel
type ServerConfig struct {
	ServerIp          string
	ServerPort        int
	ServerNetworkType string
}

func (sc *ServerConfig) Validate() error {
	if sc.ServerIp == "" {
		return conferr.ErrServerIpIsEmpty
	}
	if sc.ServerPort < 0 || sc.ServerPort > 65535 {
		return conferr.ErrServerPortIsInvalid
	}
	if sc.ServerNetworkType == "" {
		return conferr.ErrServerNetworkTypeIsEmpty
	}

	return nil
}

func (sc *ServerConfig) GetServerAddress() string {
	return sc.ServerIp + ":" + strconv.Itoa(sc.ServerPort)
}
