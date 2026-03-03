package models

import (
	conferr "insync/internal/config/errors"
	"strconv"
)

type ClientConfig struct {
	Ip          string
	Port        int
	NetworkType string

	ChunkSizeInBytes int
}

func (cc *ClientConfig) Validate() error {
	if cc.Ip == "" {
		return conferr.ErrClientIpIsEmpty
	}
	if cc.Port < 0 || cc.Port > 65535 {
		return conferr.ErrClientPortIsInvalid
	}
	if cc.NetworkType == "" {
		return conferr.ErrClientNetworkTypeIsEmpty
	}
	if cc.ChunkSizeInBytes <= 0 {
		return conferr.ErrChunkSizeIsInvalid
	}

	return nil
}

func (cc *ClientConfig) GetClientAddress() string {
	return cc.Ip + ":" + strconv.Itoa(cc.Port)
}
