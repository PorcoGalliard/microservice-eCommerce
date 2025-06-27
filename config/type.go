package config

type (
	Config struct {
		App AppConfig `yaml:"app" validate:"required"`
		Database DatabaseConfig `yaml:"database" validate:"required"`
		Redis RedisConfig `yaml:"redis" validate:"required"`
		Secret SecretConfig `yaml:"secret" validate:"required"`
	}

	AppConfig struct {
		Port string `yaml:"port" validate:"required"`
	}

	DatabaseConfig struct {
		Host string `yaml:"host" validate:"required"`
		User string `yaml:"user" validate:"required"`
		Password string `yaml:"password" validate:"required"`
		Name string `yaml:"name" validate:"required"`
		Port string `yaml:"port" validate:"required"`
	}

	RedisConfig struct {
		Host string `json:"host" validate:"required"`
		Port string `json:"port" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	SecretConfig struct {
		JWTSecret string `json:"jwt_secret" validate:"required"`
	}
)