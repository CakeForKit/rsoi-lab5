package config

var CoreConfig baseConfig

type baseConfig struct {
	Port     int          `yaml:"port,omitempty"`
	SchemaDB string       `yaml:"schemaDB,omitempty"`
	Postgres GormProperty `yaml:"postgres,omitempty"`
	Auth     AuthProperty `yaml:"auth,omitempty"`
}

// AuthProperty describes the OpenID Connect provider used by all services.
type AuthProperty struct {
	Issuer       string `yaml:"issuer,omitempty"`
	JWKSURL      string `yaml:"jwksUrl,omitempty"`
	TokenURL     string `yaml:"tokenUrl,omitempty"`
	ClientID     string `yaml:"clientId,omitempty"`
	ClientSecret string `yaml:"clientSecret,omitempty"`
	UserClaim    string `yaml:"userClaim,omitempty"`
	JWKSCacheTTL int    `yaml:"jwksCacheTtlSeconds,omitempty"`
}

type GormProperty struct {
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	Database string `yaml:"database,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	DSN      string `yaml:"dsn,omitempty"`
	ShowSql  bool   `yaml:"showSql,omitempty"`
}
