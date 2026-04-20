package models

import "fmt"

// Реализует интерфейс ConfigModel
type DbConfig struct {
	Dir      string `mapstructure:"dir"`
	FileName string `mapstructure:"file_name"`
	Pragma   string `mapstructure:"pragma"`
}

func (dc *DbConfig) Validate() error {
	if dc.Dir == "" {
		return ErrDbDirIsEmpty
	}

	if dc.FileName == "" {
		return ErrDbFileNameIsEmpty
	}

	return nil
}

func (dc *DbConfig) GetPath() string {
	return fmt.Sprintf("%s/%s", dc.Dir, dc.FileName)
}

func (dc *DbConfig) GetDsn() string {
	path := dc.GetPath()
	return fmt.Sprintf("%s%s", path, dc.Pragma)
}
