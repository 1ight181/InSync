package models

type ConnectionConfig struct {
	BaseDelaySeconds         int
	Multiplier               float64
	MaxDelaySeconds          int
	Jitter                   float64
	MinConnectTimeoutSeconds int
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
