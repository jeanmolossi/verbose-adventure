package config

type Log struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}
