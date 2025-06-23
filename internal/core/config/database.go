package config

import (
	"fmt"
	"net/url"
)

type ConnStrings interface {
	MasterConnString() string
	ReplicaConnString() string
}

var (
	_ (ConnStrings) = (*MySQL)(nil)
	_ (ConnStrings) = (*PostgreSQL)(nil)
)

type Database struct {
	MySQL      MySQL
	PostgreSQL PostgreSQL
}

// MySQL configuration

type MySQL struct {
	Database    string `envconfig:"MYSQL_DATABASE"   default:"crmcore"`
	User        string `envconfig:"MYSQL_USER"       required:"true"`
	Password    string `envconfig:"MYSQL_PASS"       required:"true"`
	Port        string `envconfig:"MYSQL_PORT"       default:"3306"`
	MasterHost  string `envconfig:"MYSQL_WRITE_HOST" required:"true"`
	ReplicaHost string `envconfig:"MYSQL_READ_HOST"  required:"true"`
}

func (c MySQL) dsn(host string) string {
	query := url.Values{
		"parseTime": []string{"true"},
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/crmcore?%s",
		c.User, c.Password, host, c.Port, query.Encode(),
	)
}

func (c MySQL) MasterConnString() string {
	return c.dsn(c.MasterHost)
}

func (c MySQL) ReplicaConnString() string {
	return c.dsn(c.ReplicaHost)
}

// PostgreSQL configuration

type PostgreSQL struct {
	Database    string `envconfig:"POSTGRES_DATABASE"   default:"crmcore"`
	User        string `envconfig:"POSTGRES_USER"       required:"true"`
	Password    string `envconfig:"POSTGRES_PASSWORD"   required:"true"`
	Port        string `envconfig:"POSTGRES_PORT"       default:"5432"`
	MasterHost  string `envconfig:"POSTGRES_WRITE_HOST" required:"true"`
	ReplicaHost string `envconfig:"POSTGRES_READ_HOST"  required:"true"`
}

func (c PostgreSQL) dsn(host string) string {
	query := url.Values{
		"sslmode": []string{"disable"},
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/crmcore?%s",
		c.User, c.Password, host, c.Port, query.Encode(),
	)
}

func (c PostgreSQL) MasterConnString() string {
	return c.dsn(c.MasterHost)
}

func (c PostgreSQL) ReplicaConnString() string {
	return c.dsn(c.ReplicaHost)
}
