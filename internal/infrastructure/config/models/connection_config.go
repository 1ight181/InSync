package models

type ConnectionConfig struct {
	BaseDelaySeconds         int     `mapstructure:"base_delay_seconds"`
	Multiplier               float64 `mapstructure:"multiplier"`
	MaxDelaySeconds          int     `mapstructure:"max_delay_seconds"`
	Jitter                   float64 `mapstructure:"jitter"`
	MinConnectTimeoutSeconds int     `mapstructure:"min_connect_timeout_seconds"`
}

func (c *ConnectionConfig) Validate() error {
	if c.BaseDelaySeconds <= 0 {
		return ErrInvalidBaseDelaySeconds
	}
	if c.Multiplier <= 0 {
		return ErrInvalidMultiplier
	}
	if c.MaxDelaySeconds <= 0 {
		return ErrInvalidMaxDelaySeconds
	}
	if c.Jitter <= 0 {
		return ErrInvalidJitter
	}
	if c.MinConnectTimeoutSeconds <= 0 {
		return ErrInvalidMinConnectTimeoutSeconds
	}

	return nil
}
