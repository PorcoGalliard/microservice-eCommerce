package config

type RedisConfig struct {
	Host string `yaml:"host" validate:"required"`
	Port string `yaml:"port" validate:"required"`
	Password string `yaml:"password" validate:"required"`
}