package connection

import "insync/internal/domain"

type IMDnsUrlResolver interface {
	Resolve(nodeName domain.NodeName) string
}
