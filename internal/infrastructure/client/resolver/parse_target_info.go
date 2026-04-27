package resolver

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/grpc/resolver"
)

var mdnsTargetRe = regexp.MustCompile(
	`(?i)^([^.]+)\.(_[^.]+)\.(_[^.]+)\.((?:[^.]+\.)+)$`,
)

func parseTargetInfo(target resolver.Target) (targetInfo, error) {
	if target.URL.Scheme != "mdns" {
		return targetInfo{}, ErrUnsupportedScheme
	}

	endpoint := strings.TrimSpace(target.URL.Path)
	endpoint = strings.Trim(endpoint, "/")
	if endpoint == "" {
		return targetInfo{}, ErrEmptyEndpoint
	}

	matches := mdnsTargetRe.FindStringSubmatch(endpoint)
	if matches == nil {
		return targetInfo{}, ErrInvalidFormat
	}

	instanceName := matches[1]
	service := matches[2]
	proto := matches[3]
	domain := matches[4]

	serviceType := fmt.Sprintf("%s.%s", service, proto)

	return targetInfo{
		instanceName: instanceName,
		serviceType:  serviceType,
		domain:       domain,
	}, nil
}
