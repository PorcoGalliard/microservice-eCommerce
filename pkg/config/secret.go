package config

type SecretConfig struct {
	JWTSecret string `yaml:"jwt_secret" validate:"required"`
}