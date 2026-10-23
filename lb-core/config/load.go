package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Load(config any) error {
	viperConfig := viper.New()
	viperConfig.AutomaticEnv()
	_ = viperConfig.BindEnv("port", "PORT")
	_ = viperConfig.BindEnv("schemaDB", "SCHEMA_DB")
	_ = viperConfig.BindEnv("postgres.dsn", "DATABASE_URL")
	_ = viperConfig.BindEnv("auth.issuer", "OIDC_ISSUER")
	_ = viperConfig.BindEnv("auth.jwksUrl", "OIDC_JWKS_URL")
	_ = viperConfig.BindEnv("auth.tokenUrl", "OIDC_TOKEN_URL")
	_ = viperConfig.BindEnv("auth.clientId", "OIDC_CLIENT_ID")
	_ = viperConfig.BindEnv("auth.clientSecret", "OIDC_CLIENT_SECRET")
	_ = viperConfig.BindEnv("auth.userClaim", "OIDC_USER_CLAIM")
	_ = viperConfig.BindEnv("auth.jwksCacheTtlSeconds", "OIDC_JWKS_CACHE_TTL_SECONDS")
	viperConfig.AddConfigPath("./config")
	viperConfig.SetConfigName("application")
	viperConfig.SetConfigType("yaml")

	if err := viperConfig.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read configuration file: %s\n", err.Error())
	}
	if err := viperConfig.Unmarshal(config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}
