package config

import "github.com/PorcoGalliard/eCommerce-Microservice/pkg/config"

type UserConfig struct {
	App config.AppConfig
	Database config.PostgreConfig
	Redis config.RedisConfig
	Secret config.SecretConfig
}