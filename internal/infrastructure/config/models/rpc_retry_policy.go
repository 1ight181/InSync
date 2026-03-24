package models

type RpcRetryPolicy struct {
	MaxAttempts           int      `mapstructure:"max_attempts"`
	InitialBackoffSeconds int      `mapstructure:"initial_backoff_seconds"`
	MaxBackoffSeconds     int      `mapstructure:"max_backoff_seconds"`
	BackoffMultiplier     float64  `mapstructure:"backoff_multiplier"`
	RetryableStatusCodes  []string `mapstructure:"retryable_status_codes"`
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
