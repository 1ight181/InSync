package models

type RpcRetryPolicy struct {
	MaxAttempts           int
	InitialBackoffSeconds int
	MaxBackoffSeconds     int
	BackoffMultiplier     float64
	RetryableStatusCodes  []string
}

func (r *RpcRetryPolicy) Validate() error {
	if r.MaxAttempts <= 0 {
		return ErrInvalidMaxAttempts
	}
	if r.InitialBackoffSeconds <= 0 {
		return ErrInvalidInitialBackoffSeconds
	}
	if r.MaxBackoffSeconds <= 0 {
		return ErrInvalidMaxBackoffSeconds
	}
	if r.BackoffMultiplier <= 0 {
		return ErrInvalidBackoffMultiplier
	}

	return nil
}
