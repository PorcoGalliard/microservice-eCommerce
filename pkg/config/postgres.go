package config

type PostgreConfig struct {
	Host string `yaml:"host" validate:"required"`
	User string `yaml:"user" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Name string `yaml:"name" validate:"required"`
	Port string `yaml:"port" validate:"required"`
}