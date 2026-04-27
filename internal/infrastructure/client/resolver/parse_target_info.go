package resolver

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/resolver"
)

func parseTargetInfo(target resolver.Target) (targetInfo, error) {
	url := target.URL

	if url.Scheme != "mdns" {
		return targetInfo{}, ErrInvalidScheme
	}

	endpoint := strings.Trim(url.Path, "/")

	splited := strings.Split(endpoint, ".")
	if len(splited) != 5 {
		return targetInfo{}, ErrInvalidEndpointFormat
	}

	serviceName := fmt.Sprintf("%s.%s", splited[0], splited[1])
	if serviceName == "" {
		return targetInfo{}, ErrMissingServiceName
	}

	instanceName := fmt.Sprintf("%s.%s", splited[2], splited[3])

	domain := splited[4]

	return targetInfo{
		instanceName: instanceName,
		serviceName:  serviceName,
		domain:       domain,
	}, nil
}
