package models

import "fmt"

// Реализует интерфейс ConfigModel
type DbConfig struct {
	Path   string `mapstructure:"path"`
	Pragma string `mapstructure:"pragma"`
}

func (dc *DbConfig) Validate() error {
	if dc.Path == "" {
		return ErrDbDsnIsEmpty
	}

	return nil
}

func (dc *DbConfig) GetDsn() string {
	return fmt.Sprintf("%s%s", dc.Path, dc.Pragma)
}
