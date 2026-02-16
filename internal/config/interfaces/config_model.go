package config

type IConfigModel interface {
	Validate() error
}
