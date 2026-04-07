package models

type FileManagerConfig struct {
	TempDir string `yaml:"temp_dir"`
}

func (c *FileManagerConfig) Validate() error {
	if c.TempDir == "" {
		return ErrTempDirIsEmpty
	}

	return nil
}
