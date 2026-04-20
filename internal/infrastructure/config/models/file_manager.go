package models

type FileManagerConfig struct {
	TempDir string `mapstructure:"temp_dir"`
}

func (c *FileManagerConfig) Validate() error {
	if c.TempDir == "" {
		return ErrTempDirIsEmpty
	}

	return nil
}
