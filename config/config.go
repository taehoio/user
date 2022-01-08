package config

type Config interface {
	Setting() Setting
}

type DefaultConfig struct {
	setting Setting
}

func (c *DefaultConfig) Setting() Setting {
	return c.setting
}

func NewConfig(setting Setting) Config {
	return &DefaultConfig{
		setting: setting,
	}
}
