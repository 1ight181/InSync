package resolver

import (
	"strings"

	"google.golang.org/grpc/resolver"
)

func parseTargetInfo(target resolver.Target) (targetInfo, error) {
	url := target.URL

	if url.Scheme != "mdns" {
		return targetInfo{}, ErrInvalidScheme
	}

	serviceName := url.Host
	if serviceName == "" {
		return targetInfo{}, ErrMissingServiceName
	}

	endpoint := strings.TrimPrefix(url.Path, "/")
	if endpoint == "" {
		return targetInfo{}, ErrMissingEndpoint
	}

	lastDotIndex := strings.LastIndexByte(endpoint, '.')
	if lastDotIndex == -1 {
		return targetInfo{}, ErrInvalidEndpointFormat
	}

	instanceName := endpoint[:lastDotIndex]
	domain := endpoint[lastDotIndex+1:]

	return targetInfo{
		instanceName: instanceName,
		serviceName:  serviceName,
		domain:       domain,
	}, nil
}
