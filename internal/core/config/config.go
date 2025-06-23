package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Database Database
	Log      Log
	Session  Session

	BaseURL string `envconfig:"BASE_URL" required:"false" default:"localhost:8080"`

	SAMLKeyPath  string `envconfig:"SAML_KEY_PATH"  required:"true"`
	SAMLCertPath string `envconfig:"SAML_CERT_PATH" required:"true"`
}

func New() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
