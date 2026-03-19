package interfaces

type IMDnsServer interface {
	Start() error
	Stop() error
}
