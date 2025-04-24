package config

import (
	"fmt"
	"net/url"

	"github.com/kelseyhightower/envconfig"
)

type MySQLConfig struct {
	Database  string `envconfig:"MYSQL_DATABASE" default:"crmcore"`
	User      string `envconfig:"MYSQL_USER" required:"true"`
	Password  string `envconfig:"MYSQL_PASS" required:"true"`
	Port      string `envconfig:"MYSQL_PORT" default:"3306"`
	WriteHost string `envconfig:"MYSQL_WRITE_HOST" required:"true"`
	ReadHost  string `envconfig:"MYSQL_READ_HOST" required:"true"`
}

func (c MySQLConfig) dsn(host string) string {
	query := url.Values{
		"parseTime": []string{"true"},
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/crmcore?%s",
		c.User, c.Password, host, c.Port, query.Encode(),
	)
}

func (c MySQLConfig) WriteDSN() string {
	return c.dsn(c.WriteHost)
}

func (c MySQLConfig) ReadDSN() string {
	return c.dsn(c.ReadHost)
}

type PGConfig struct {
	Database  string `envconfig:"POSTGRES_DATABASE" default:"crmcore"`
	User      string `envconfig:"POSTGRES_USER" required:"true"`
	Password  string `envconfig:"POSTGRES_PASSWORD" required:"true"`
	Port      string `envconfig:"POSTGRES_PORT" default:"5432"`
	WriteHost string `envconfig:"POSTGRES_WRITE_HOST" required:"true"`
	ReadHost  string `envconfig:"POSTGRES_READ_HOST" required:"true"`
}

func (c PGConfig) dsn(host string) string {
	query := url.Values{
		"sslmode": []string{"disabled"},
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/crmcore?%s",
		c.User, c.Password, host, c.Port, query.Encode(),
	)
}

func (c PGConfig) WriteDSN() string {
	return c.dsn(c.WriteHost)
}

func (c PGConfig) ReadDSN() string {
	return c.dsn(c.ReadHost)
}

type Config struct {
	MySQLConfig MySQLConfig
	PGConfig    PGConfig
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
}

func New() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
